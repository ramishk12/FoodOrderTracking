package handlers

import (
	"net/http"
	"strconv"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/gin-gonic/gin"
)

func GetItems(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, name, description, price, category, available, created_at, updated_at
		FROM items WHERE available = true ORDER BY category, name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var i models.Item
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.Price, &i.Category, &i.Available, &i.CreatedAt, &i.UpdatedAt); err == nil {
			items = append(items, i)
		}
	}

	c.JSON(http.StatusOK, items)
}

func GetItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var i models.Item
	err = database.DB.QueryRow(`
		SELECT id, name, description, price, category, available, created_at, updated_at
		FROM items WHERE id = $1
	`, id).Scan(&i.ID, &i.Name, &i.Description, &i.Price, &i.Category, &i.Available, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, i)
}

func CreateItem(c *gin.Context) {
	var input models.Item
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if input.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return
	}

	if input.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
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

	input.ID = id
	c.JSON(http.StatusCreated, input)
}

func UpdateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var input models.Item
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if input.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return
	}

	if input.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
		return
	}

	_, err = database.DB.Exec(`
		UPDATE items SET name = $1, description = $2, price = $3, category = $4, available = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`, input.Name, input.Description, input.Price, input.Category, input.Available, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item updated"})
}

func DeactivateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	_, err = database.DB.Exec("UPDATE items SET available = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deactivated"})
}
