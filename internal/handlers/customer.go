package handlers

import (
	"net/http"
	"strconv"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/gin-gonic/gin"
)

func GetCustomers(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, name, phone, email, address, created_at, updated_at
		FROM customers ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Address, &c.CreatedAt, &c.UpdatedAt); err == nil {
			customers = append(customers, c)
		}
	}

	c.JSON(http.StatusOK, customers)
}

func GetCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var customer models.Customer
	err = database.DB.QueryRow(`
		SELECT id, name, phone, email, address, created_at, updated_at
		FROM customers WHERE id = $1
	`, id).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Email, &customer.Address, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func CreateCustomer(c *gin.Context) {
	var input models.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := database.DB.QueryRow(`
		INSERT INTO customers (name, phone, email, address)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, input.Name, input.Phone, input.Email, input.Address).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	input.ID = id
	c.JSON(http.StatusCreated, input)
}

func UpdateCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var input models.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = database.DB.Exec(`
		UPDATE customers SET name = $1, phone = $2, email = $3, address = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`, input.Name, input.Phone, input.Email, input.Address, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer updated"})
}

func DeleteCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM customers WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted"})
}
