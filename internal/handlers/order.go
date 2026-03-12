package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/gin-gonic/gin"
)

// validStatuses is the single source of truth for allowed order statuses.
var validStatuses = map[string]bool{
	"pending":   true,
	"preparing": true,
	"ready":     true,
	"delivered": true,
	"cancelled": true,
}

// orderQuery is the shared SELECT clause used across all order queries.
const orderQuery = `
	SELECT o.id, o.customer_id, o.delivery_address, o.status, o.total_amount,
	       o.notes, o.payment_method, o.scheduled_date, o.created_at, o.updated_at,
	       COALESCE(c.name, ''), COALESCE(c.phone, '')
	FROM orders o
	LEFT JOIN customers c ON o.customer_id = c.id`

// scanOrder scans a single order row into a models.Order.
func scanOrder(row interface {
	Scan(...any) error
}) (models.Order, error) {
	var o models.Order
	err := row.Scan(
		&o.ID, &o.CustomerID, &o.DeliveryAddress, &o.Status, &o.TotalAmount,
		&o.Notes, &o.PaymentMethod, &o.ScheduledDate, &o.CreatedAt, &o.UpdatedAt,
		&o.CustomerName, &o.CustomerPhone,
	)
	return o, err
}

// populateOrderItems fetches all order items for a slice of orders in a single
// query (avoids N+1) and attaches them to the corresponding orders.
func populateOrderItems(orders []models.Order) error {
	if len(orders) == 0 {
		return nil
	}

	// Build an IN clause with all order IDs.
	ids := make([]any, len(orders))
	placeholders := ""
	for i, o := range orders {
		ids[i] = o.ID
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
	}

	rows, err := database.DB.Query(fmt.Sprintf(`
		SELECT oi.id, oi.order_id, oi.item_id, COALESCE(i.name, ''),
		       oi.quantity, oi.unit_price, oi.subtotal
		FROM order_items oi
		LEFT JOIN items i ON oi.item_id = i.id
		WHERE oi.order_id IN (%s)
		ORDER BY oi.id
	`, placeholders), ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Index orders by ID for O(1) lookup, and pre-initialize each order's
	// items to an empty slice so JSON always encodes [] rather than null.
	orderMap := make(map[int]*models.Order, len(orders))
	for i := range orders {
		orders[i].OrderItems = make([]models.OrderItem, 0)
		orderMap[orders[i].ID] = &orders[i]
	}

	for rows.Next() {
		var oi models.OrderItem
		if err := rows.Scan(&oi.ID, &oi.OrderID, &oi.ItemID, &oi.ItemName, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal); err != nil {
			log.Printf("Error scanning order item: %v", err)
			continue
		}
		if o, ok := orderMap[oi.OrderID]; ok {
			o.OrderItems = append(o.OrderItems, oi)
		}
	}
	return rows.Err()
}

// normalizeScheduledDate converts a scheduled date pointer to UTC in-place.
func normalizeScheduledDate(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}

// validateOrderStatus returns an error response if the status is invalid.
func validateOrderStatus(c *gin.Context, status string) bool {
	if status != "" && !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return false
	}
	return true
}

// customerExists checks whether a customer ID exists in the database.
func customerExists(customerID int) (bool, error) {
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM customers WHERE id = $1)", customerID).Scan(&exists)
	return exists, err
}

// GetOrders returns all orders with their items.
func GetOrders(c *gin.Context) {
	rows, err := database.DB.Query(orderQuery + `
		ORDER BY COALESCE(o.scheduled_date, o.created_at) DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			log.Printf("Error scanning order: %v", err)
			continue
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := populateOrderItems(orders); err != nil {
		log.Printf("Error populating order items: %v", err)
	}

	c.JSON(http.StatusOK, orders)
}

// GetScheduledOrders returns upcoming non-completed orders within a date window.
func GetScheduledOrders(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}

	now := time.Now().UTC()
	endDate := now.AddDate(0, 0, days)

	rows, err := database.DB.Query(orderQuery+`
		WHERE o.scheduled_date IS NOT NULL
		  AND o.scheduled_date BETWEEN $1 AND $2
		  AND o.status NOT IN ('delivered', 'cancelled')
		ORDER BY o.scheduled_date ASC
	`, now, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			log.Printf("Error scanning scheduled order: %v", err)
			continue
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := populateOrderItems(orders); err != nil {
		log.Printf("Error populating order items: %v", err)
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrdersByCustomer returns the 10 most recent orders for a customer.
func GetOrdersByCustomer(c *gin.Context) {
	customerID, err := strconv.Atoi(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	rows, err := database.DB.Query(orderQuery+`
		WHERE o.customer_id = $1
		ORDER BY o.created_at DESC
		LIMIT 10
	`, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			log.Printf("Error scanning customer order: %v", err)
			continue
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := populateOrderItems(orders); err != nil {
		log.Printf("Error populating order items: %v", err)
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrder returns a single order by ID.
func GetOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	row := database.DB.QueryRow(orderQuery+` WHERE o.id = $1`, id)
	o, err := scanOrder(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	orders := []models.Order{o}
	if err := populateOrderItems(orders); err != nil {
		log.Printf("Error populating order items for order %d: %v", id, err)
	}
	o.OrderItems = orders[0].OrderItems

	c.JSON(http.StatusOK, o)
}

// createOrderInput is the expected request body for creating an order.
type createOrderInput struct {
	CustomerID      int        `json:"customer_id"`
	DeliveryAddress string     `json:"delivery_address"`
	Status          string     `json:"status"`
	Notes           string     `json:"notes"`
	PaymentMethod   string     `json:"payment_method"`
	ScheduledDate   *time.Time `json:"scheduled_date"`
	Items           []struct {
		ItemID   int `json:"item_id"`
		Quantity int `json:"quantity"`
	} `json:"items"`
}

// updateOrderInput is the expected request body for updating an order.
type updateOrderInput struct {
	CustomerID      int        `json:"customer_id"`
	DeliveryAddress string     `json:"delivery_address"`
	Status          string     `json:"status"`
	TotalAmount     float64    `json:"total_amount"`
	Notes           string     `json:"notes"`
	PaymentMethod   string     `json:"payment_method"`
	ScheduledDate   *time.Time `json:"scheduled_date"`
	Items           []struct {
		ItemID   int `json:"item_id"`
		Quantity int `json:"quantity"`
	} `json:"items"`
}

// CreateOrder creates a new order and its associated items in a transaction.
func CreateOrder(c *gin.Context) {
	var input createOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "pending"
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = "cash"
	}
	if !validateOrderStatus(c, input.Status) {
		return
	}
	if len(input.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one item is required"})
		return
	}
	if input.CustomerID != 0 {
		exists, err := customerExists(input.CustomerID)
		if err != nil || !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Customer not found"})
			return
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// Validate items and compute total.
	var totalAmount float64
	itemPrices := make(map[int]float64, len(input.Items))
	for _, item := range input.Items {
		if item.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than 0"})
			return
		}
		var price float64
		if err := tx.QueryRow("SELECT price FROM items WHERE id = $1", item.ItemID).Scan(&price); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid item ID: %d", item.ItemID)})
			return
		}
		itemPrices[item.ItemID] = price
		totalAmount += price * float64(item.Quantity)
	}

	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes, payment_method, scheduled_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, input.CustomerID, input.DeliveryAddress, input.Status, totalAmount,
		input.Notes, input.PaymentMethod, normalizeScheduledDate(input.ScheduledDate),
	).Scan(&orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, item := range input.Items {
		price := itemPrices[item.ItemID]
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal)
			VALUES ($1, $2, $3, $4, $5)
		`, orderID, item.ItemID, item.Quantity, price, price*float64(item.Quantity))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": orderID, "status": input.Status, "total_amount": totalAmount})
}

// UpdateOrder updates an existing order's details and optionally its items.
func UpdateOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var input updateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.PaymentMethod == "" {
		input.PaymentMethod = "cash"
	}
	if !validateOrderStatus(c, input.Status) {
		return
	}
	if input.CustomerID > 0 {
		exists, err := customerExists(input.CustomerID)
		if err != nil || !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Customer not found"})
			return
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE orders
		SET customer_id = $1, delivery_address = $2, status = $3, total_amount = $4,
		    notes = $5, payment_method = $6, scheduled_date = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
	`, input.CustomerID, input.DeliveryAddress, input.Status, input.TotalAmount,
		input.Notes, input.PaymentMethod, normalizeScheduledDate(input.ScheduledDate), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(input.Items) > 0 {
		if _, err = tx.Exec("DELETE FROM order_items WHERE order_id = $1", id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, item := range input.Items {
			if item.Quantity <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than 0"})
				return
			}
			var unitPrice float64
			if err = tx.QueryRow("SELECT price FROM items WHERE id = $1", item.ItemID).Scan(&unitPrice); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid item ID: %d", item.ItemID)})
				return
			}
			_, err = tx.Exec(`
				INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal)
				VALUES ($1, $2, $3, $4, $5)
			`, id, item.ItemID, item.Quantity, unitPrice, unitPrice*float64(item.Quantity))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order updated"})
}

// DeleteOrder removes an order and all its items in a transaction.
func DeleteOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM order_items WHERE order_id = $1", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err = tx.Exec("DELETE FROM orders WHERE id = $1", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted"})
}