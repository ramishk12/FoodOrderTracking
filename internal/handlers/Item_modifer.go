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

const modifierQuery = `SELECT id, name, price_adjustment FROM item_modifiers`

func scanModifier(row interface {
	Scan(...any) error
}) (models.ItemModifier, error) {
	var m models.ItemModifier
	err := row.Scan(&m.ID, &m.Name, &m.PriceAdjustment)
	return m, err
}

// GetModifiers returns all item modifiers ordered by name.
func GetModifiers(c *gin.Context) {
	rows, err := database.DB.Query(modifierQuery + ` ORDER BY name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	modifiers := make([]models.ItemModifier, 0)
	for rows.Next() {
		m, err := scanModifier(rows)
		if err != nil {
			log.Printf("Error scanning modifier: %v", err)
			continue
		}
		modifiers = append(modifiers, m)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, modifiers)
}

// CreateModifier inserts a new item modifier.
func CreateModifier(c *gin.Context) {
	var input models.ItemModifier
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	var id int
	if err := database.DB.QueryRow(`
		INSERT INTO item_modifiers (name, price_adjustment)
		VALUES ($1, $2) RETURNING id
	`, input.Name, input.PriceAdjustment).Scan(&id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := scanModifier(database.DB.QueryRow(modifierQuery+` WHERE id = $1`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// UpdateModifier updates a modifier's name and price adjustment.
func UpdateModifier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid modifier ID"})
		return
	}

	var input models.ItemModifier
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	result, err := database.DB.Exec(`
		UPDATE item_modifiers
		SET name = $1, price_adjustment = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, input.Name, input.PriceAdjustment, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Modifier not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Modifier updated"})
}

// DeleteModifier removes a modifier by ID.
// Existing order_item_modifiers referencing it will have modifier_id set to NULL (ON DELETE SET NULL).
func DeleteModifier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid modifier ID"})
		return
	}

	result, err := database.DB.Exec(`DELETE FROM item_modifiers WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Modifier not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Modifier deleted"})
}

// GetModifier returns a single modifier by ID.
func GetModifier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid modifier ID"})
		return
	}

	m, err := scanModifier(database.DB.QueryRow(modifierQuery+` WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Modifier not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}
