package handlers

import (
	"net/http"
	"strconv"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/gin-gonic/gin"
)

func GetOrders(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT o.id, o.customer_id, o.delivery_address, o.status, o.total_amount, o.items, o.notes, o.created_at, o.updated_at,
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
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.DeliveryAddress, &o.Status, &o.TotalAmount, &o.Items, &o.Notes, &o.CreatedAt, &o.UpdatedAt, &o.CustomerName, &o.CustomerPhone); err == nil {
			orders = append(orders, o)
		}
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
		SELECT o.id, o.customer_id, o.delivery_address, o.status, o.total_amount, o.items, o.notes, o.created_at, o.updated_at,
		       COALESCE(c.name, ''), COALESCE(c.phone, '')
		FROM orders o
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE o.id = $1
	`, id).Scan(&o.ID, &o.CustomerID, &o.DeliveryAddress, &o.Status, &o.TotalAmount, &o.Items, &o.Notes, &o.CreatedAt, &o.UpdatedAt, &o.CustomerName, &o.CustomerPhone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, o)
}

func CreateOrder(c *gin.Context) {
	var input struct {
		CustomerID      int     `json:"customer_id"`
		DeliveryAddress string  `json:"delivery_address"`
		Status          string  `json:"status"`
		TotalAmount     float64 `json:"total_amount"`
		Items           string  `json:"items"`
		Notes           string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = "pending"
	}

	var id int
	err := database.DB.QueryRow(`
		INSERT INTO orders (customer_id, delivery_address, status, total_amount, items, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, input.CustomerID, input.DeliveryAddress, input.Status, input.TotalAmount, input.Items, input.Notes).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "status": input.Status})
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
		Items           string  `json:"items"`
		Notes           string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = database.DB.Exec(`
		UPDATE orders SET customer_id = $1, delivery_address = $2, status = $3, total_amount = $4, items = $5, notes = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`, input.CustomerID, input.DeliveryAddress, input.Status, input.TotalAmount, input.Items, input.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
