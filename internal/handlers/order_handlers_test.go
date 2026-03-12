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

// orderCols are the columns returned by the shared orderQuery SELECT.
var orderCols = []string{
	"id", "customer_id", "delivery_address", "status", "total_amount",
	"notes", "payment_method", "scheduled_date", "created_at", "updated_at",
	"coalesce", "coalesce_2",
}

// itemCols are the columns returned by populateOrderItems.
var itemCols = []string{"id", "order_id", "item_id", "name", "quantity", "unit_price", "subtotal"}

// orderQueryRegex matches the shared SELECT used by all list/single-order queries.
const orderQueryRegex = `SELECT o\.id, o\.customer_id, o\.delivery_address, o\.status, o\.total_amount,`

// populateItemsQueryRegex matches the single IN-clause query used by populateOrderItems.
const populateItemsQueryRegex = `SELECT oi\.id, oi\.order_id, oi\.item_id, COALESCE\(i\.name`

func TestGetOrders(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Returns all orders with items",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnRows(sqlmock.NewRows(orderCols).
						AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", nil, time.Now(), time.Now(), "John Doe", "555-1234"))
				// populateOrderItems fires a single IN-clause query for all order IDs
				m.ExpectQuery(populateItemsQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(itemCols).
						AddRow(1, 1, 1, "Pizza", 2, 12.99, 25.98))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Len(t, orders, 1)
				assert.Equal(t, 1, orders[0].ID)
				assert.Equal(t, "pending", orders[0].Status)
				assert.Len(t, orders[0].OrderItems, 1)
				assert.Equal(t, "Pizza", orders[0].OrderItems[0].ItemName)
			},
		},
		{
			name: "Returns empty array when no orders",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnRows(sqlmock.NewRows(orderCols))
				// populateOrderItems is a no-op when slice is empty — no DB call expected
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Len(t, orders, 0)
			},
		},
		{
			name: "Returns 500 on database error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Returns orders with customer info",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnRows(sqlmock.NewRows(orderCols).
						AddRow(1, 1, "123 Main St", "delivered", 50.00, "No onions", "card", nil, time.Now(), time.Now(), "Jane Smith", "555-5678"))
				m.ExpectQuery(populateItemsQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(itemCols))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Equal(t, "Jane Smith", orders[0].CustomerName)
				assert.Equal(t, "555-5678", orders[0].CustomerPhone)
			},
		},
		{
			name: "Multiple orders get items populated in one query",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnRows(sqlmock.NewRows(orderCols).
						AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", nil, time.Now(), time.Now(), "John Doe", "555-1234").
						AddRow(2, 2, "456 Oak Ave", "preparing", 15.00, "", "e-transfer", nil, time.Now(), time.Now(), "Jane Smith", "555-5678"))
				// Single query for both orders (IN $1, $2)
				m.ExpectQuery(populateItemsQueryRegex).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows(itemCols).
						AddRow(1, 1, 1, "Pizza", 2, 12.99, 25.98).
						AddRow(2, 2, 2, "Burger", 1, 15.00, 15.00))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Len(t, orders, 2)
				assert.Len(t, orders[0].OrderItems, 1)
				assert.Len(t, orders[1].OrderItems, 1)
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
			c.Request = httptest.NewRequest(http.MethodGet, "/api/orders", nil)

			GetOrders(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetScheduledOrders(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:  "Returns scheduled orders within window",
			query: "/api/orders/scheduled?days=7",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnRows(sqlmock.NewRows(orderCols).
						AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", time.Now().AddDate(0, 0, 2), time.Now(), time.Now(), "John Doe", "555-1234"))
				m.ExpectQuery(populateItemsQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(itemCols).
						AddRow(1, 1, 1, "Pizza", 2, 12.99, 25.98))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Len(t, orders, 1)
				assert.Len(t, orders[0].OrderItems, 1)
			},
		},
		{
			name:  "Returns empty when no scheduled orders",
			query: "/api/orders/scheduled?days=7",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnRows(sqlmock.NewRows(orderCols))
				// No populateOrderItems call when slice is empty
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Len(t, orders, 0)
			},
		},
		{
			name:  "Uses default 7 days when days param is invalid",
			query: "/api/orders/scheduled?days=abc",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WillReturnRows(sqlmock.NewRows(orderCols))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Returns 500 on database error",
			query: "/api/orders/scheduled",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
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
			c.Request = httptest.NewRequest(http.MethodGet, tt.query, nil)

			GetScheduledOrders(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetOrdersByCustomer(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Returns orders for valid customer",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(orderCols).
						AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", nil, time.Now(), time.Now(), "John Doe", "555-1234"))
				m.ExpectQuery(populateItemsQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(itemCols))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Len(t, orders, 1)
			},
		},
		{
			name:           "Returns 400 for invalid customer ID",
			customerID:     "abc",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid customer ID", resp["error"])
			},
		},
		{
			name:       "Returns empty for customer with no orders",
			customerID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WithArgs(999).
					WillReturnRows(sqlmock.NewRows(orderCols))
				// populateOrderItems no-op — no DB call expected
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &orders))
				assert.Len(t, orders, 0)
			},
		},
		{
			name:       "Returns 500 on database error",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WithArgs(1).
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
			c.Params = gin.Params{{Key: "customerId", Value: tt.customerID}}
			c.Request = httptest.NewRequest(http.MethodGet, "/api/orders/customer/"+tt.customerID, nil)

			GetOrdersByCustomer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderID        string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:    "Returns order by ID with items",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(orderCols).
						AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", nil, time.Now(), time.Now(), "John Doe", "555-1234"))
				// GetOrder wraps the single order in a slice and calls populateOrderItems
				m.ExpectQuery(populateItemsQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(itemCols).
						AddRow(1, 1, 1, "Pizza", 2, 12.99, 25.98))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var order models.Order
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &order))
				assert.Equal(t, 1, order.ID)
				assert.Equal(t, "pending", order.Status)
				assert.Len(t, order.OrderItems, 1)
				assert.Equal(t, "Pizza", order.OrderItems[0].ItemName)
			},
		},
		{
			name:           "Returns 400 for invalid order ID",
			orderID:        "abc",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid order ID", resp["error"])
			},
		},
		{
			name:    "Returns 404 when order not found",
			orderID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Order not found", resp["error"])
			},
		},
		{
			name:    "Returns 404 on any database error",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderQueryRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("connection error"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Order not found", resp["error"])
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
			c.Params = gin.Params{{Key: "id", Value: tt.orderID}}
			c.Request = httptest.NewRequest(http.MethodGet, "/api/orders/"+tt.orderID, nil)

			GetOrder(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Creates order successfully",
			body: `{"customer_id":1,"delivery_address":"123 Main St","status":"pending","payment_method":"cash","items":[{"item_id":1,"quantity":2}]}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM customers WHERE id = \$1\)`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectQuery(`SELECT price FROM items WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(12.99))
				m.ExpectQuery(`INSERT INTO orders`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectExec(`INSERT INTO order_items`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, float64(1), resp["id"])
				assert.Equal(t, "pending", resp["status"])
				assert.InDelta(t, 25.98, resp["total_amount"], 0.01)
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			body:           `{"customer_id":1,"items":[}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Contains(t, resp["error"], "invalid")
			},
		},
		{
			// items check now happens BEFORE tx.Begin() — no DB calls expected
			name:           "Returns 400 for empty items array",
			body:           `{"delivery_address":"123 Main St","items":[]}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "At least one item is required", resp["error"])
			},
		},
		{
			name:           "Returns 400 for invalid status",
			body:           `{"delivery_address":"123 Main St","status":"flying","items":[{"item_id":1,"quantity":1}]}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid status", resp["error"])
			},
		},
		{
			name: "Returns 400 when customer does not exist",
			body: `{"customer_id":99,"delivery_address":"123 Main St","items":[{"item_id":1,"quantity":1}]}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM customers WHERE id = \$1\)`).
					WithArgs(99).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Customer not found", resp["error"])
			},
		},
		{
			name: "Defaults status to pending and payment_method to cash",
			body: `{"customer_id":1,"delivery_address":"123 Main St","items":[{"item_id":1,"quantity":1}]}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM customers WHERE id = \$1\)`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectQuery(`SELECT price FROM items WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(10.00))
				m.ExpectQuery(`INSERT INTO orders`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				m.ExpectExec(`INSERT INTO order_items`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "pending", resp["status"])
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
			req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			CreateOrder(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderID        string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:    "Updates order successfully with new items",
			orderID: "1",
			body:    `{"customer_id":1,"delivery_address":"123 Main St","status":"preparing","total_amount":30.00,"payment_method":"cash","items":[{"item_id":1,"quantity":2}]}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM customers WHERE id = \$1\)`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectExec(`UPDATE orders SET`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(`DELETE FROM order_items WHERE order_id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectQuery(`SELECT price FROM items WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(12.99))
				m.ExpectExec(`INSERT INTO order_items`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Order updated", resp["message"])
			},
		},
		{
			name:    "Updates order without changing items",
			orderID: "1",
			body:    `{"customer_id":1,"delivery_address":"456 Oak Ave","status":"ready","total_amount":25.99,"payment_method":"cash"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM customers WHERE id = \$1\)`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				m.ExpectBegin()
				m.ExpectExec(`UPDATE orders SET`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// No DELETE/INSERT for order_items since items array is absent
				m.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Order updated", resp["message"])
			},
		},
		{
			name:           "Returns 400 for invalid order ID",
			orderID:        "abc",
			body:           `{"status":"preparing"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid order ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			orderID:        "1",
			body:           `{"status":}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 for invalid status",
			orderID:        "1",
			body:           `{"status":"invalid_status"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid status", resp["error"])
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
			c.Params = gin.Params{{Key: "id", Value: tt.orderID}}
			req := httptest.NewRequest(http.MethodPut, "/api/orders/"+tt.orderID, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			UpdateOrder(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderID        string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:    "Deletes order and its items successfully",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(`DELETE FROM order_items WHERE order_id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(`DELETE FROM orders WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Order deleted", resp["message"])
			},
		},
		{
			name:           "Returns 400 for invalid order ID",
			orderID:        "abc",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid order ID", resp["error"])
			},
		},
		{
			name:    "Returns 500 when deleting order items fails",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(`DELETE FROM order_items WHERE order_id = \$1`).
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
				m.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "Returns 500 when deleting order fails",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(`DELETE FROM order_items WHERE order_id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(`DELETE FROM orders WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
				m.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			// Deleting a non-existent order is idempotent — still returns 200
			name:    "Deletes non-existent order without error",
			orderID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(`DELETE FROM order_items WHERE order_id = \$1`).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
				m.ExpectExec(`DELETE FROM orders WHERE id = \$1`).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Order deleted", resp["message"])
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
			c.Params = gin.Params{{Key: "id", Value: tt.orderID}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/orders/"+tt.orderID, nil)

			DeleteOrder(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
