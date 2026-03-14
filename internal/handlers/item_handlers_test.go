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

	"food-order-tracking/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ── Column/query constants ────────────────────────────────────────────────────

var itemQueryCols = []string{"id", "name", "description", "price", "category", "available", "created_at", "updated_at"}

const (
	itemQueryRegex       = `SELECT id, name, description, price, category, available, created_at, updated_at\s+FROM items`
	insertItemQueryRegex = `INSERT INTO items`
	updateItemExecRegex  = `UPDATE items SET`
	deactivateExecRegex  = `UPDATE items SET available = false`
	activateExecRegex    = `UPDATE items SET available = true`
)

// ── Row helpers ───────────────────────────────────────────────────────────────

func itemRow(id int, name, description string, price float64, category string, available bool) *sqlmock.Rows {
	return sqlmock.NewRows(itemQueryCols).
		AddRow(id, name, description, price, category, available, time.Now(), time.Now())
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestGetItems(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Returns all items",
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
				assert.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
			},
		},
		{
			name: "Returns 500 on database error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp["error"])
			},
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
		{
			name: "Continues scanning after a bad row",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
					WillReturnRows(sqlmock.NewRows(itemQueryCols).
						AddRow("not-an-int", "Bad Row", "", 0, "", false, time.Now(), time.Now()).
						AddRow(2, "Good Row", "", 9.99, "Main", true, time.Now(), time.Now()))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var items []models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
				assert.Len(t, items, 1)
				assert.Equal(t, "Good Row", items[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
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
					WillReturnRows(itemRow(1, "Pizza", "Delicious pizza", 12.99, "Main", true))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var item models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
				assert.Equal(t, 1, item.ID)
				assert.Equal(t, "Pizza", item.Name)
				assert.Equal(t, 12.99, item.Price)
				assert.Equal(t, "Main", item.Category)
				assert.True(t, item.Available)
				assert.False(t, item.CreatedAt.IsZero())
			},
		},
		{
			name:           "Returns 400 for non-numeric ID",
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
			name:           "Returns 400 for zero ID",
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
			name:           "Returns 400 for negative ID",
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
			name:   "Returns 500 on database error",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(itemQueryRegex).
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
				c.Request = httptest.NewRequest(http.MethodGet, "/api/items/"+tt.itemID, nil)

				GetItem(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
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
				m.ExpectQuery(insertItemQueryRegex).
					WithArgs("Pizza", "Delicious pizza", 12.99, "Main", true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				// Handler fetches full record after INSERT to return DB-generated timestamps.
				m.ExpectQuery(itemQueryRegex).
					WithArgs(1).
					WillReturnRows(itemRow(1, "Pizza", "Delicious pizza", 12.99, "Main", true))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var item models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
				assert.Equal(t, 1, item.ID)
				assert.Equal(t, "Pizza", item.Name)
				assert.Equal(t, 12.99, item.Price)
				assert.False(t, item.CreatedAt.IsZero(), "CreatedAt should be set by the database")
				assert.False(t, item.UpdatedAt.IsZero(), "UpdatedAt should be set by the database")
			},
		},
		{
			name: "Creates unavailable item",
			body: `{"name":"Burger","description":"Tasty burger","price":8.99,"category":"Main","available":false}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertItemQueryRegex).
					WithArgs("Burger", "Tasty burger", 8.99, "Main", false).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				m.ExpectQuery(itemQueryRegex).
					WithArgs(2).
					WillReturnRows(itemRow(2, "Burger", "Tasty burger", 8.99, "Main", false))
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
			name: "Trims whitespace from string fields before storing",
			body: `{"name":"  Pizza  ","description":" Delicious ","price":12.99,"category":" Main ","available":true}`,
			setupMock: func(m sqlmock.Sqlmock) {
				// WithArgs must be the trimmed values — untrimmed would fail.
				m.ExpectQuery(insertItemQueryRegex).
					WithArgs("Pizza", "Delicious", 12.99, "Main", true).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
				m.ExpectQuery(itemQueryRegex).
					WithArgs(3).
					WillReturnRows(itemRow(3, "Pizza", "Delicious", 12.99, "Main", true))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var item models.Item
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
				assert.Equal(t, "Pizza", item.Name)
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
				assert.NotEmpty(t, resp["error"])
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
			name:           "Returns 400 for blank whitespace name",
			body:           `{"name":"   ","price":12.99,"category":"Main"}`,
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
			name: "Returns 500 on database error",
			body: `{"name":"Pizza","price":12.99,"category":"Main"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertItemQueryRegex).
					WillReturnError(fmt.Errorf("database error"))
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
				m.ExpectExec(updateItemExecRegex).
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
			name:   "Updates item to unavailable",
			itemID: "2",
			body:   `{"name":"Burger","description":"Out of stock","price":8.99,"category":"Main","available":false}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateItemExecRegex).
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
			name:   "Trims whitespace from string fields before storing",
			itemID: "1",
			body:   `{"name":"  Salad  ","description":" Fresh ","price":6.99,"category":" Side ","available":true}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateItemExecRegex).
					WithArgs("Salad", "Fresh", 6.99, "Side", true, 1).
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
			name:           "Returns 400 for non-numeric ID",
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
			name:           "Returns 400 for zero ID",
			itemID:         "0",
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
			name:           "Returns 400 for negative ID",
			itemID:         "-1",
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
			name:   "Returns 404 when item does not exist",
			itemID: "999",
			body:   `{"name":"Ghost","price":9.99,"category":"Main"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateItemExecRegex).
					WithArgs("Ghost", "", 9.99, "Main", false, 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Item not found", resp["error"])
			},
		},
		{
			name:   "Returns 500 on database error",
			itemID: "1",
			body:   `{"name":"Pizza","price":12.99,"category":"Main"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateItemExecRegex).
					WithArgs("Pizza", "", 12.99, "Main", false, 1).
					WillReturnError(fmt.Errorf("database error"))
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
				m.ExpectExec(deactivateExecRegex).
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
			name:           "Returns 400 for non-numeric ID",
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
			name:           "Returns 400 for zero ID",
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
			name:           "Returns 400 for negative ID",
			itemID:         "-3",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid item ID", resp["error"])
			},
		},
		{
			name:   "Returns 404 when item does not exist",
			itemID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(deactivateExecRegex).
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
		{
			name:   "Returns 500 on database error",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(deactivateExecRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
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
				c.Request = httptest.NewRequest(http.MethodPatch, "/api/items/"+tt.itemID+"/deactivate", nil)

				DeactivateItem(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

func TestActivateItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "Activates item successfully",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(activateExecRegex).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Item activated", resp["message"])
			},
		},
		{
			name:           "Returns 400 for non-numeric ID",
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
			name:           "Returns 400 for zero ID",
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
			name:           "Returns 400 for negative ID",
			itemID:         "-2",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid item ID", resp["error"])
			},
		},
		{
			name:   "Returns 404 when item does not exist",
			itemID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(activateExecRegex).
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
		{
			name:   "Returns 500 on database error",
			itemID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(activateExecRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
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
				c.Request = httptest.NewRequest(http.MethodPatch, "/api/items/"+tt.itemID+"/activate", nil)

				ActivateItem(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}
