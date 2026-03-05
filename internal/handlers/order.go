package handlers

import (
	"log"
	"net/http"
	"strconv"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/gin-gonic/gin"
)

func GetOrders(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT o.id, o.customer_id, o.delivery_address, o.status, o.total_amount, o.notes, o.created_at, o.updated_at,
		       COALESCE(c.name, ''), COALESCE(c.phone, '')
		FROM orders o
		LEFT JOIN customers c ON o.customer_id = c.id
		ORDER BY o.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.DeliveryAddress, &o.Status, &o.TotalAmount, &o.Notes, &o.CreatedAt, &o.UpdatedAt, &o.CustomerName, &o.CustomerPhone); err == nil {
			orders = append(orders, o)
		}
	}

	// Get order items for each order
	for i := range orders {
		itemRows, err := database.DB.Query(`
			SELECT oi.id, oi.order_id, oi.item_id, i.name, oi.quantity, oi.unit_price, oi.subtotal
			FROM order_items oi
			JOIN items i ON oi.item_id = i.id
			WHERE oi.order_id = $1
		`, orders[i].ID)
		if err != nil {
			log.Printf("Error fetching items for order %d: %v", orders[i].ID, err)
			continue
		}

		var orderItems []models.OrderItem
		for itemRows.Next() {
			var oi models.OrderItem
			if err := itemRows.Scan(&oi.ID, &oi.OrderID, &oi.ItemID, &oi.ItemName, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal); err == nil {
				orderItems = append(orderItems, oi)
			}
		}
		itemRows.Close()

		orders[i].OrderItems = orderItems
	}

	c.JSON(http.StatusOK, orders)
}

func GetOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var o models.Order
	err = database.DB.QueryRow(`
		SELECT o.id, o.customer_id, o.delivery_address, o.status, o.total_amount, o.notes, o.created_at, o.updated_at,
		       COALESCE(c.name, ''), COALESCE(c.phone, '')
		FROM orders o
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE o.id = $1
	`, id).Scan(&o.ID, &o.CustomerID, &o.DeliveryAddress, &o.Status, &o.TotalAmount, &o.Notes, &o.CreatedAt, &o.UpdatedAt, &o.CustomerName, &o.CustomerPhone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	itemRows, err := database.DB.Query(`
		SELECT oi.id, oi.order_id, oi.item_id, COALESCE(i.name, ''), oi.quantity, oi.unit_price, oi.subtotal
		FROM order_items oi
		LEFT JOIN items i ON oi.item_id = i.id
		WHERE oi.order_id = $1
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer itemRows.Close()

	var orderItems []models.OrderItem
	for itemRows.Next() {
		var oi models.OrderItem
		if err := itemRows.Scan(&oi.ID, &oi.OrderID, &oi.ItemID, &oi.ItemName, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal); err == nil {
			orderItems = append(orderItems, oi)
		}
	}
	o.OrderItems = orderItems

	c.JSON(http.StatusOK, o)
}

func CreateOrder(c *gin.Context) {
	var input struct {
		CustomerID      int    `json:"customer_id"`
		DeliveryAddress string `json:"delivery_address"`
		Status          string `json:"status"`
		Notes           string `json:"notes"`
		Items           []struct {
			ItemID   int `json:"item_id"`
			Quantity int `json:"quantity"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "pending"
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var totalAmount float64
	for _, item := range input.Items {
		var price float64
		err := tx.QueryRow("SELECT price FROM items WHERE id = $1", item.ItemID).Scan(&price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item"})
			return
		}
		totalAmount += price * float64(item.Quantity)
	}

	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, input.CustomerID, input.DeliveryAddress, input.Status, totalAmount, input.Notes).Scan(&orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, item := range input.Items {
		var price float64
		tx.QueryRow("SELECT price FROM items WHERE id = $1", item.ItemID).Scan(&price)
		subtotal := price * float64(item.Quantity)

		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal)
			VALUES ($1, $2, $3, $4, $5)
		`, orderID, item.ItemID, item.Quantity, price, subtotal)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"id": orderID, "status": input.Status, "total_amount": totalAmount})
}

func UpdateOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var input struct {
		CustomerID      int     `json:"customer_id"`
		DeliveryAddress string  `json:"delivery_address"`
		Status          string  `json:"status"`
		TotalAmount     float64 `json:"total_amount"`
		Notes           string  `json:"notes"`
		Items           []struct {
			ItemID   int `json:"item_id"`
			Quantity int `json:"quantity"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// Update order details
	_, err = tx.Exec(`
		UPDATE orders SET customer_id = $1, delivery_address = $2, status = $3, total_amount = $4, notes = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`, input.CustomerID, input.DeliveryAddress, input.Status, input.TotalAmount, input.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If items provided, update order_items
	if len(input.Items) > 0 {
		// Delete existing order items
		_, err = tx.Exec("DELETE FROM order_items WHERE order_id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert new order items
		for _, item := range input.Items {
			if item.Quantity <= 0 {
				continue
			}
			var unitPrice float64
			tx.QueryRow("SELECT price FROM items WHERE id = $1", item.ItemID).Scan(&unitPrice)
			subtotal := unitPrice * float64(item.Quantity)

			_, err = tx.Exec(`
				INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal)
				VALUES ($1, $2, $3, $4, $5)
			`, id, item.ItemID, item.Quantity, unitPrice, subtotal)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Order updated"})
}

func DeleteOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted"})
}
