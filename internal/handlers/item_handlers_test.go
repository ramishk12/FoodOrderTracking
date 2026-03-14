package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// itemQueryCols are the columns returned by itemQuery SELECT.
var itemQueryCols = []string{"id", "name", "description", "price", "category", "available", "created_at", "updated_at"}

// itemQueryRegex matches the shared SELECT used by GetItems and GetItem.
const itemQueryRegex = `SELECT id, name, description, price, category, available, created_at, updated_at\s+FROM items`

func TestGetItems(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Returns all available items",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WillReturnRows(sqlmock.NewRows(itemQueryCols).
						AddRow(1, "Pizza", "Delicious pizza", 12.99, "Main", true, time.Now(), time.Now()).
						AddRow(2, "Burger", "Tasty burger", 8.99, "Main", true, time.Now(), time.Now()).
						AddRow(3, "Salad", "Fresh salad", 6.99, "Side", true, time.Now(), time.Now()))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var items []models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
				assert.Len(t, items, 3)
				assert.Equal(t, "Pizza", items[0].Name)
				assert.Equal(t, "Burger", items[1].Name)
				assert.Equal(t, "Salad", items[2].Name)
			},
		},
		{
			name: "Returns empty array when no items",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WillReturnRows(sqlmock.NewRows(itemQueryCols))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var items []models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
				assert.Len(t, items, 0)
			},
		},
		{
			name: "Returns 500 on database error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Returns items ordered by category then name",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WillReturnRows(sqlmock.NewRows(itemQueryCols).
						AddRow(1, "Coke", "Beverage", 2.99, "Drinks", true, time.Now(), time.Now()).
						AddRow(2, "Pasta", "Italian pasta", 10.99, "Main", true, time.Now(), time.Now()).
						AddRow(3, "Pizza", "Main dish", 12.99, "Main", true, time.Now(), time.Now()))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var items []models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
				assert.Len(t, items, 3)
				assert.Equal(t, "Drinks", items[0].Category)
				assert.Equal(t, "Main", items[1].Category)
				assert.Equal(t, "Main", items[2].Category)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			assert.NoError(t, err)
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/items", nil)

			GetItems(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "Returns item by ID",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(itemQueryCols).
						AddRow(1, "Pizza", "Delicious pizza", 12.99, "Main", true, time.Now(), time.Now()))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var item models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
				assert.Equal(t, "Pizza", item.Name)
				assert.Equal(t, 12.99, item.Price)
			},
		},
		{
			name:           "Returns 400 for invalid item ID",
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
			name:   "Returns 404 when item not found",
			itemID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Item not found", resp["error"])
			},
		},
		{
			name:   "Returns item with all fields populated",
			itemID: "2",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows(itemQueryCols).
						AddRow(2, "Burger", "Tasty beef burger", 8.99, "Main", true,
							time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var item models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
				assert.Equal(t, 2, item.ID)
				assert.Equal(t, "Burger", item.Name)
				assert.Equal(t, "Tasty beef burger", item.Description)
				assert.Equal(t, 8.99, item.Price)
				assert.Equal(t, "Main", item.Category)
				assert.True(t, item.Available)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			assert.NoError(t, err)
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.itemID}}
			c.Request = httptest.NewRequest(http.MethodGet, "/api/items/"+tt.itemID, nil)

			GetItem(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateItem(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Creates item successfully",
			body: `{"name":"Pizza","description":"Delicious pizza","price":12.99,"category":"Main","available":true}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO items`).
					WithArgs("Pizza", "Delicious pizza", 12.99, "Main", true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery(itemQueryRegex).
					WillReturnRows(sqlmock.NewRows(itemQueryCols).
						AddRow(1, "Pizza", "Delicious pizza", 12.99, "Main", true, time.Now(), time.Now()))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var item models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
				assert.Equal(t, 1, item.ID)
				assert.Equal(t, "Pizza", item.Name)
				assert.Equal(t, 12.99, item.Price)
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			body:           `{"name":"Pizza","price":}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Contains(t, resp["error"], "invalid")
			},
		},
		{
			name:           "Returns 400 for missing name",
			body:           `{"price":12.99,"category":"Main"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Name is required", resp["error"])
			},
		},
		{
			name:           "Returns 400 for missing category",
			body:           `{"name":"Pizza","price":12.99}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Category is required", resp["error"])
			},
		},
		{
			name:           "Returns 400 for zero price",
			body:           `{"name":"Pizza","price":0,"category":"Main"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Price must be greater than 0", resp["error"])
			},
		},
		{
			name:           "Returns 400 for negative price",
			body:           `{"name":"Pizza","price":-5.99,"category":"Main"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Price must be greater than 0", resp["error"])
			},
		},
		{
			name: "Creates unavailable item",
			body: `{"name":"Burger","description":"Tasty burger","price":8.99,"category":"Main","available":false}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO items`).
					WithArgs("Burger", "Tasty burger", 8.99, "Main", false).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				m.ExpectQuery(itemQueryRegex).
					WillReturnRows(sqlmock.NewRows(itemQueryCols).
						AddRow(2, "Burger", "Tasty burger", 8.99, "Main", false, time.Now(), time.Now()))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var item models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
				assert.Equal(t, 2, item.ID)
				assert.False(t, item.Available)
			},
		},
		{
			name: "Returns 500 on database error",
			body: `{"name":"Pizza","price":12.99,"category":"Main"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO items`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			assert.NoError(t, err)
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodPost, "/api/items", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			CreateItem(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "Updates item successfully",
			itemID: "1",
			body:   `{"name":"Pizza","description":"Updated pizza","price":14.99,"category":"Main","available":true}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items SET`).
					WithArgs("Pizza", "Updated pizza", 14.99, "Main", true, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Item updated", resp["message"])
			},
		},
		{
			name:           "Returns 400 for invalid item ID",
			itemID:         "abc",
			body:           `{"name":"Pizza","price":12.99,"category":"Main"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid item ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			itemID:         "1",
			body:           `{"name":"Pizza","price":}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 for missing name",
			itemID:         "1",
			body:           `{"price":12.99,"category":"Main"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Name is required", resp["error"])
			},
		},
		{
			name:           "Returns 400 for missing category",
			itemID:         "1",
			body:           `{"name":"Pizza","price":12.99}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Category is required", resp["error"])
			},
		},
		{
			name:           "Returns 400 for zero price",
			itemID:         "1",
			body:           `{"name":"Pizza","price":0,"category":"Main"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Price must be greater than 0", resp["error"])
			},
		},
		{
			name:           "Returns 400 for negative price",
			itemID:         "1",
			body:           `{"name":"Pizza","price":-5.99,"category":"Main"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Price must be greater than 0", resp["error"])
			},
		},
		{
			name:   "Updates item to unavailable",
			itemID: "2",
			body:   `{"name":"Burger","description":"Out of stock","price":8.99,"category":"Main","available":false}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items SET`).
					WithArgs("Burger", "Out of stock", 8.99, "Main", false, 2).
					WillReturnResult(sqlmock.NewResult(2, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Item updated", resp["message"])
			},
		},
		{
			name:   "Returns 500 on database error",
			itemID: "1",
			body:   `{"name":"Pizza","price":12.99,"category":"Main"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items SET`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			assert.NoError(t, err)
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.itemID}}
			req := httptest.NewRequest(http.MethodPut, "/api/items/"+tt.itemID, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			UpdateItem(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeactivateItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "Deactivates item successfully",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items SET available = false`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Item deactivated", resp["message"])
			},
		},
		{
			name:           "Returns 400 for invalid item ID",
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
			name:   "Returns 500 on database error",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items SET available = false`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			// Deactivating a non-existent item now returns 404
			name:   "Deactivates non-existent item without error",
			itemID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items SET available = false`).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Item not found", resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			assert.NoError(t, err)
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.itemID}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/items/"+tt.itemID, nil)

			DeactivateItem(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
