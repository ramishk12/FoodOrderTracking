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

// itemQuery is the shared SELECT clause used across all item queries.
const itemQuery = `
	SELECT id, name, description, price, category, available, created_at, updated_at
	FROM items`

// scanItem scans a single item row into a models.Item.
func scanItem(row interface {
	Scan(...any) error
}) (models.Item, error) {
	var i models.Item
	err := row.Scan(&i.ID, &i.Name, &i.Description, &i.Price, &i.Category, &i.Available, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

// trimItem returns a copy of input with all string fields whitespace-trimmed.
func trimItem(input models.Item) models.Item {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.TrimSpace(input.Category)
	return input
}

// validateItem returns false and writes an error response if the item input is invalid.
func validateItem(c *gin.Context, input models.Item) bool {
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return false
	}
	if input.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return false
	}
	if input.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
		return false
	}
	return true
}

// GetItems returns all items ordered by category and name.
func GetItems(c *gin.Context) {
	rows, err := database.DB.Query(itemQuery + ` ORDER BY category, name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]models.Item, 0)
	for rows.Next() {
		i, err := scanItem(rows)
		if err != nil {
			log.Printf("Error scanning item: %v", err)
			continue
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := populateItemModifiers(items); err != nil {
		log.Printf("Error populating item modifiers: %v", err)
	}

	c.JSON(http.StatusOK, items)
}

// GetItem returns a single item by ID.
func GetItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	row := database.DB.QueryRow(itemQuery+` WHERE id = $1`, id)
	i, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := []models.Item{i}
	if err := populateItemModifiers(items); err != nil {
		log.Printf("Error populating modifiers for item %d: %v", id, err)
	}

	c.JSON(http.StatusOK, items[0])
}

// CreateItem inserts a new menu item.
func CreateItem(c *gin.Context) {
	var input models.Item
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input = trimItem(input)
	if !validateItem(c, input) {
		return
	}

	var id int
	err := database.DB.QueryRow(`
		INSERT INTO items (name, description, price, category, available)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, input.Name, input.Description, input.Price, input.Category, input.Available).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := scanItem(database.DB.QueryRow(itemQuery+` WHERE id = $1`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// UpdateItem updates an existing menu item by ID.
func UpdateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var input models.Item
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input = trimItem(input)
	if !validateItem(c, input) {
		return
	}

	result, err := database.DB.Exec(`
		UPDATE items
		SET name = $1, description = $2, price = $3, category = $4, available = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`, input.Name, input.Description, input.Price, input.Category, input.Available, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item updated"})
}

// DeactivateItem marks an item as unavailable without deleting it.
func DeactivateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	result, err := database.DB.Exec(
		`UPDATE items SET available = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deactivated"})
}

// ActivateItem marks an item as available.
func ActivateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	result, err := database.DB.Exec(
		`UPDATE items SET available = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item activated"})
}
