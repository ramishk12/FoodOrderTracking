package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/gin-gonic/gin"
)

// customerQuery is the shared SELECT clause used across all customer queries.
const customerQuery = `
	SELECT id, name, phone, email, address, created_at, updated_at
	FROM customers`

// scanCustomer scans a single customer row into a models.Customer.
func scanCustomer(row interface {
	Scan(...any) error
}) (models.Customer, error) {
	var c models.Customer
	err := row.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Address, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

// trimCustomer returns a copy of input with all string fields whitespace-trimmed.
func trimCustomer(input models.Customer) models.Customer {
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.TrimSpace(input.Email)
	input.Address = strings.TrimSpace(input.Address)
	return input
}

// validateCustomer returns false and writes an error response if the input is invalid.
func validateCustomer(c *gin.Context, input models.Customer) bool {
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return false
	}
	return true
}

// GetCustomers returns all customers ordered by creation date descending.
func GetCustomers(c *gin.Context) {
	rows, err := database.DB.Query(customerQuery + ` ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	customers := make([]models.Customer, 0)
	for rows.Next() {
		cust, err := scanCustomer(rows)
		if err != nil {
			log.Printf("Error scanning customer: %v", err)
			continue
		}
		customers = append(customers, cust)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}

// GetCustomer returns a single customer by ID.
func GetCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	row := database.DB.QueryRow(customerQuery+` WHERE id = $1`, id)
	customer, err := scanCustomer(row)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// CreateCustomer inserts a new customer record.
func CreateCustomer(c *gin.Context) {
	var input models.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input = trimCustomer(input)
	if !validateCustomer(c, input) {
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

	created, err := scanCustomer(database.DB.QueryRow(customerQuery+` WHERE id = $1`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// UpdateCustomer updates an existing customer's details by ID.
func UpdateCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var input models.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input = trimCustomer(input)
	if !validateCustomer(c, input) {
		return
	}

	result, err := database.DB.Exec(`
		UPDATE customers
		SET name = $1, phone = $2, email = $3, address = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`, input.Name, input.Phone, input.Email, input.Address, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer updated"})
}

// DeleteCustomer removes a customer by ID, rejecting the request if the
// customer has any associated orders.
func DeleteCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var orderCount int
	if err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM orders WHERE customer_id = $1`, id,
	).Scan(&orderCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if orderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete customer with existing orders"})
		return
	}

	delResult, err := database.DB.Exec(`DELETE FROM customers WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := delResult.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted"})
}
