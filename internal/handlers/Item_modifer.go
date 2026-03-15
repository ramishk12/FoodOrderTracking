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

// modifierQuery is the shared SELECT for item_modifiers.
const modifierQuery = `SELECT id, item_id, name, price_adjustment FROM item_modifiers`

func scanModifier(row interface {
	Scan(...any) error
}) (models.ItemModifier, error) {
	var m models.ItemModifier
	err := row.Scan(&m.ID, &m.ItemID, &m.Name, &m.PriceAdjustment)
	return m, err
}

// populateItemModifiers fetches modifiers for a slice of items in a single
// query and attaches them to the corresponding items.
func populateItemModifiers(items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	ids := make([]any, len(items))
	placeholders := ""
	for i, item := range items {
		ids[i] = item.ID
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "$" + strconv.Itoa(i+1)
	}

	rows, err := database.DB.Query(
		"SELECT id, item_id, name, price_adjustment FROM item_modifiers WHERE item_id IN ("+placeholders+") ORDER BY name",
		ids...,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Index items by ID for O(1) attachment.
	itemMap := make(map[int]*models.Item, len(items))
	for i := range items {
		itemMap[items[i].ID] = &items[i]
	}

	for rows.Next() {
		m, err := scanModifier(rows)
		if err != nil {
			log.Printf("Error scanning item modifier: %v", err)
			continue
		}
		if item, ok := itemMap[m.ItemID]; ok {
			item.Modifiers = append(item.Modifiers, m)
		}
	}
	return rows.Err()
}

// GetItemModifiers returns all modifiers for a specific item.
func GetItemModifiers(c *gin.Context) {
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	rows, err := database.DB.Query(modifierQuery+` WHERE item_id = $1 ORDER BY name`, itemID)
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

// CreateItemModifier adds a modifier to a specific item.
func CreateItemModifier(c *gin.Context) {
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
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

	var id int
	if err := database.DB.QueryRow(`
		INSERT INTO item_modifiers (item_id, name, price_adjustment)
		VALUES ($1, $2, $3) RETURNING id
	`, itemID, input.Name, input.PriceAdjustment).Scan(&id); err != nil {
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

// UpdateItemModifier updates a modifier's name and price adjustment.
func UpdateItemModifier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("modifierId"))
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

// DeleteItemModifier removes a modifier.
// Existing order_item_modifiers referencing it will have modifier_id set to NULL.
func DeleteItemModifier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("modifierId"))
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

// GetItemModifier returns a single modifier by its ID.
func GetItemModifier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("modifierId"))
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
