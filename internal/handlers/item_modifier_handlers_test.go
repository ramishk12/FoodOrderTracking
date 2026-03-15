package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"food-order-tracking/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ── Column/query constants ────────────────────────────────────────────────────

var modifierCols = []string{"id", "item_id", "name", "price_adjustment"}

const (
	modifierListQueryRegex   = `SELECT id, item_id, name, price_adjustment FROM item_modifiers WHERE item_id`
	modifierSingleQueryRegex = `SELECT id, item_id, name, price_adjustment FROM item_modifiers WHERE id`
	insertModifierQueryRegex = `INSERT INTO item_modifiers`
	updateModifierExecRegex  = `UPDATE item_modifiers\s+SET`
	deleteModifierExecRegex  = `DELETE FROM item_modifiers WHERE id`
)

// ── Row helpers ───────────────────────────────────────────────────────────────

func modifierRow(id, itemID int, name string, priceAdj float64) *sqlmock.Rows {
	return sqlmock.NewRows(modifierCols).AddRow(id, itemID, name, priceAdj)
}

// ── TestGetItemModifiers ──────────────────────────────────────────────────────

func TestGetItemModifiers(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "Returns all modifiers for an item",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(modifierListQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(modifierCols).
						AddRow(1, 1, "Extra Cheese", 1.50).
						AddRow(2, 1, "No Onions", 0.00))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var mods []models.ItemModifier
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &mods))
				assert.Len(t, mods, 2)
				assert.Equal(t, "Extra Cheese", mods[0].Name)
				assert.Equal(t, 1.50, mods[0].PriceAdjustment)
				assert.Equal(t, "No Onions", mods[1].Name)
				assert.Equal(t, 0.00, mods[1].PriceAdjustment)
			},
		},
		{
			name:   "Returns empty array when item has no modifiers",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(modifierListQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(modifierCols))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
			},
		},
		{
			name:           "Returns 400 for non-numeric item ID",
			itemID:         "abc",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid item ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for zero item ID",
			itemID:         "0",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid item ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for negative item ID",
			itemID:         "-1",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid item ID", resp["error"])
			},
		},
		{
			name:   "Returns 500 on database error",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(modifierListQueryRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{{Key: "id", Value: tt.itemID}}
				c.Request = httptest.NewRequest(http.MethodGet, "/api/items/"+tt.itemID+"/modifiers", nil)

				GetItemModifiers(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

// ── TestGetItemModifier ───────────────────────────────────────────────────────

func TestGetItemModifier(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		modifierID     string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Returns modifier by ID",
			itemID:     "1",
			modifierID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(modifierSingleQueryRegex).
					WithArgs(1).
					WillReturnRows(modifierRow(1, 1, "Extra Cheese", 1.50))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var mod models.ItemModifier
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &mod))
				assert.Equal(t, 1, mod.ID)
				assert.Equal(t, 1, mod.ItemID)
				assert.Equal(t, "Extra Cheese", mod.Name)
				assert.Equal(t, 1.50, mod.PriceAdjustment)
			},
		},
		{
			name:           "Returns 400 for non-numeric modifier ID",
			itemID:         "1",
			modifierID:     "abc",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid modifier ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for zero modifier ID",
			itemID:         "1",
			modifierID:     "0",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid modifier ID", resp["error"])
			},
		},
		{
			name:       "Returns 404 when modifier not found",
			itemID:     "1",
			modifierID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(modifierSingleQueryRegex).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Modifier not found", resp["error"])
			},
		},
		{
			name:       "Returns 500 on database error",
			itemID:     "1",
			modifierID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(modifierSingleQueryRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{
					{Key: "id", Value: tt.itemID},
					{Key: "modifierId", Value: tt.modifierID},
				}
				c.Request = httptest.NewRequest(http.MethodGet,
					"/api/items/"+tt.itemID+"/modifiers/"+tt.modifierID, nil)

				GetItemModifier(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

// ── TestCreateItemModifier ────────────────────────────────────────────────────

func TestCreateItemModifier(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "Creates modifier with positive price adjustment",
			itemID: "1",
			body:   `{"name":"Extra Cheese","price_adjustment":1.50}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertModifierQueryRegex).
					WithArgs(1, "Extra Cheese", 1.50).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				// Handler fetches full record after INSERT.
				m.ExpectQuery(modifierSingleQueryRegex).
					WithArgs(1).
					WillReturnRows(modifierRow(1, 1, "Extra Cheese", 1.50))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var mod models.ItemModifier
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &mod))
				assert.Equal(t, 1, mod.ID)
				assert.Equal(t, 1, mod.ItemID)
				assert.Equal(t, "Extra Cheese", mod.Name)
				assert.Equal(t, 1.50, mod.PriceAdjustment)
			},
		},
		{
			name:   "Creates modifier with zero price adjustment",
			itemID: "1",
			body:   `{"name":"No Onions","price_adjustment":0}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertModifierQueryRegex).
					WithArgs(1, "No Onions", 0.0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				m.ExpectQuery(modifierSingleQueryRegex).
					WithArgs(2).
					WillReturnRows(modifierRow(2, 1, "No Onions", 0.0))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var mod models.ItemModifier
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &mod))
				assert.Equal(t, "No Onions", mod.Name)
				assert.Equal(t, 0.0, mod.PriceAdjustment)
			},
		},
		{
			name:   "Creates modifier with negative price adjustment",
			itemID: "1",
			body:   `{"name":"Remove Bacon","price_adjustment":-0.50}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertModifierQueryRegex).
					WithArgs(1, "Remove Bacon", -0.50).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
				m.ExpectQuery(modifierSingleQueryRegex).
					WithArgs(3).
					WillReturnRows(modifierRow(3, 1, "Remove Bacon", -0.50))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var mod models.ItemModifier
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &mod))
				assert.Equal(t, -0.50, mod.PriceAdjustment)
			},
		},
		{
			name:   "Trims whitespace from name",
			itemID: "1",
			body:   `{"name":"  Extra Sauce  ","price_adjustment":0.50}`,
			setupMock: func(m sqlmock.Sqlmock) {
				// WithArgs must match the trimmed name.
				m.ExpectQuery(insertModifierQueryRegex).
					WithArgs(1, "Extra Sauce", 0.50).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))
				m.ExpectQuery(modifierSingleQueryRegex).
					WithArgs(4).
					WillReturnRows(modifierRow(4, 1, "Extra Sauce", 0.50))
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Returns 400 for non-numeric item ID",
			itemID:         "abc",
			body:           `{"name":"Extra Cheese","price_adjustment":1.50}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid item ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for zero item ID",
			itemID:         "0",
			body:           `{"name":"Extra Cheese","price_adjustment":1.50}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 for invalid JSON",
			itemID:         "1",
			body:           `{invalid`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 when name is missing",
			itemID:         "1",
			body:           `{"price_adjustment":1.50}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Name is required", resp["error"])
			},
		},
		{
			name:           "Returns 400 when name is blank whitespace",
			itemID:         "1",
			body:           `{"name":"   ","price_adjustment":1.50}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Name is required", resp["error"])
			},
		},
		{
			name:   "Returns 500 on database error",
			itemID: "1",
			body:   `{"name":"Extra Cheese","price_adjustment":1.50}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertModifierQueryRegex).
					WillReturnError(fmt.Errorf("unique constraint violation"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{{Key: "id", Value: tt.itemID}}
				req := httptest.NewRequest(http.MethodPost,
					"/api/items/"+tt.itemID+"/modifiers",
					strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				CreateItemModifier(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

// ── TestUpdateItemModifier ────────────────────────────────────────────────────

func TestUpdateItemModifier(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		modifierID     string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Updates modifier successfully",
			itemID:     "1",
			modifierID: "1",
			body:       `{"name":"Extra Cheese","price_adjustment":2.00}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateModifierExecRegex).
					WithArgs("Extra Cheese", 2.00, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Modifier updated", resp["message"])
			},
		},
		{
			name:       "Updates price adjustment to zero",
			itemID:     "1",
			modifierID: "2",
			body:       `{"name":"No Onions","price_adjustment":0}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateModifierExecRegex).
					WithArgs("No Onions", 0.0, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Trims whitespace from name before storing",
			itemID:     "1",
			modifierID: "1",
			body:       `{"name":"  Mushrooms  ","price_adjustment":1.00}`,
			setupMock: func(m sqlmock.Sqlmock) {
				// WithArgs must be the trimmed name.
				m.ExpectExec(updateModifierExecRegex).
					WithArgs("Mushrooms", 1.00, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Returns 400 for non-numeric modifier ID",
			itemID:         "1",
			modifierID:     "abc",
			body:           `{"name":"Extra Cheese","price_adjustment":1.50}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid modifier ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for zero modifier ID",
			itemID:         "1",
			modifierID:     "0",
			body:           `{"name":"Extra Cheese","price_adjustment":1.50}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 for invalid JSON",
			itemID:         "1",
			modifierID:     "1",
			body:           `{invalid`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 when name is missing",
			itemID:         "1",
			modifierID:     "1",
			body:           `{"price_adjustment":1.50}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Name is required", resp["error"])
			},
		},
		{
			name:       "Returns 404 when modifier does not exist",
			itemID:     "1",
			modifierID: "999",
			body:       `{"name":"Ghost","price_adjustment":1.00}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateModifierExecRegex).
					WithArgs("Ghost", 1.00, 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Modifier not found", resp["error"])
			},
		},
		{
			name:       "Returns 500 on database error",
			itemID:     "1",
			modifierID: "1",
			body:       `{"name":"Extra Cheese","price_adjustment":1.50}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateModifierExecRegex).
					WithArgs("Extra Cheese", 1.50, 1).
					WillReturnError(fmt.Errorf("deadlock detected"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{
					{Key: "id", Value: tt.itemID},
					{Key: "modifierId", Value: tt.modifierID},
				}
				req := httptest.NewRequest(http.MethodPut,
					"/api/items/"+tt.itemID+"/modifiers/"+tt.modifierID,
					strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				UpdateItemModifier(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

// ── TestDeleteItemModifier ────────────────────────────────────────────────────

func TestDeleteItemModifier(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		modifierID     string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Deletes modifier successfully",
			itemID:     "1",
			modifierID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(deleteModifierExecRegex).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Modifier deleted", resp["message"])
			},
		},
		{
			name:           "Returns 400 for non-numeric modifier ID",
			itemID:         "1",
			modifierID:     "abc",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid modifier ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for zero modifier ID",
			itemID:         "1",
			modifierID:     "0",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 for negative modifier ID",
			itemID:         "1",
			modifierID:     "-3",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Returns 404 when modifier does not exist",
			itemID:     "1",
			modifierID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(deleteModifierExecRegex).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Modifier not found", resp["error"])
			},
		},
		{
			name:       "Returns 500 on database error",
			itemID:     "1",
			modifierID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(deleteModifierExecRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("lock timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{
					{Key: "id", Value: tt.itemID},
					{Key: "modifierId", Value: tt.modifierID},
				}
				c.Request = httptest.NewRequest(http.MethodDelete,
					"/api/items/"+tt.itemID+"/modifiers/"+tt.modifierID, nil)

				DeleteItemModifier(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}
