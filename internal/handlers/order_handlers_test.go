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
				ordersRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"}).
					AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", nil, time.Now(), time.Now(), "John Doe", "555-1234")
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WillReturnRows(ordersRow)

				itemsRow := sqlmock.NewRows([]string{"id", "order_id", "item_id", "name", "quantity", "unit_price", "subtotal"}).
					AddRow(1, 1, 1, "Pizza", 2, 12.99, 25.98)
				m.ExpectQuery("SELECT oi\\.id, oi\\.order_id, oi\\.item_id, i\\.name, oi\\.quantity, oi\\.unit_price, oi\\.subtotal FROM order_items oi JOIN items i ON oi\\.item_id = i\\.id WHERE oi\\.order_id = \\$1").
					WithArgs(1).
					WillReturnRows(itemsRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, 1)
				assert.Equal(t, 1, orders[0].ID)
				assert.Equal(t, "pending", orders[0].Status)
				assert.Len(t, orders[0].OrderItems, 1)
			},
		},
		{
			name: "Returns empty array when no orders",
			setupMock: func(m sqlmock.Sqlmock) {
				ordersRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"})
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WillReturnRows(ordersRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, 0)
			},
		},
		{
			name: "Returns 500 on database error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
		{
			name: "Returns orders with customer info",
			setupMock: func(m sqlmock.Sqlmock) {
				ordersRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"}).
					AddRow(1, 1, "123 Main St", "delivered", 50.00, "No onions", "card", nil, time.Now(), time.Now(), "Jane Smith", "555-5678")
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WillReturnRows(ordersRow)

				itemsRow := sqlmock.NewRows([]string{"id", "order_id", "item_id", "name", "quantity", "unit_price", "subtotal"})
				m.ExpectQuery("SELECT oi\\.id, oi\\.order_id, oi\\.item_id, i\\.name, oi\\.quantity, oi\\.unit_price, oi\\.subtotal FROM order_items oi JOIN items i ON oi\\.item_id = i\\.id WHERE oi\\.order_id = \\$1").
					WithArgs(1).
					WillReturnRows(itemsRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Equal(t, "Jane Smith", orders[0].CustomerName)
				assert.Equal(t, "555-5678", orders[0].CustomerPhone)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
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
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:  "Returns scheduled orders",
			query: "/api/orders/scheduled?days=7",
			setupMock: func(m sqlmock.Sqlmock, query string) {
				ordersRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"}).
					AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", time.Now().AddDate(0, 0, 2), time.Now(), time.Now(), "John Doe", "555-1234")
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WillReturnRows(ordersRow)

				itemsRow := sqlmock.NewRows([]string{"id", "order_id", "item_id", "name", "quantity", "unit_price", "subtotal"}).
					AddRow(1, 1, 1, "Pizza", 2, 12.99, 25.98)
				m.ExpectQuery("SELECT oi\\.id, oi\\.order_id, oi\\.item_id, i\\.name, oi\\.quantity, oi\\.unit_price, oi\\.subtotal FROM order_items oi JOIN items i ON oi\\.item_id = i\\.id WHERE oi\\.order_id = \\$1").
					WithArgs(1).
					WillReturnRows(itemsRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, 1)
			},
		},
		{
			name:  "Returns empty when no scheduled orders",
			query: "/api/orders/scheduled?days=7",
			setupMock: func(m sqlmock.Sqlmock, query string) {
				ordersRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"})
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WillReturnRows(ordersRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, 0)
			},
		},
		{
			name:  "Returns 500 on database error",
			query: "/api/orders/scheduled",
			setupMock: func(m sqlmock.Sqlmock, query string) {
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock, tt.query)

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
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Returns orders for valid customer",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				ordersRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"}).
					AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", nil, time.Now(), time.Now(), "John Doe", "555-1234")
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WithArgs(1).
					WillReturnRows(ordersRow)

				itemsRow := sqlmock.NewRows([]string{"id", "order_id", "item_id", "name", "quantity", "unit_price", "subtotal"})
				m.ExpectQuery("SELECT oi\\.id, oi\\.order_id, oi\\.item_id, i\\.name, oi\\.quantity, oi\\.unit_price, oi\\.subtotal FROM order_items oi JOIN items i ON oi\\.item_id = i\\.id WHERE oi\\.order_id = \\$1").
					WithArgs(1).
					WillReturnRows(itemsRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, 1)
			},
		},
		{
			name:           "Returns 400 for invalid customer ID",
			customerID:     "abc",
			setupMock:      func(m sqlmock.Sqlmock, id string) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid customer ID", resp["error"])
			},
		},
		{
			name:       "Returns empty for customer with no orders",
			customerID: "999",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				ordersRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"})
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WithArgs(999).
					WillReturnRows(ordersRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var orders []models.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, 0)
			},
		},
		{
			name:       "Returns 500 on database error",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock, tt.customerID)

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
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:    "Returns order by ID",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				orderRow := sqlmock.NewRows([]string{"id", "customer_id", "delivery_address", "status", "total_amount", "notes", "payment_method", "scheduled_date", "created_at", "updated_at", "coalesce", "coalesce_2"}).
					AddRow(1, 1, "123 Main St", "pending", 25.99, "", "cash", nil, time.Now(), time.Now(), "John Doe", "555-1234")
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WithArgs(1).
					WillReturnRows(orderRow)

				itemsRow := sqlmock.NewRows([]string{"id", "order_id", "item_id", "name", "quantity", "unit_price", "subtotal"}).
					AddRow(1, 1, 1, "Pizza", 2, 12.99, 25.98)
				m.ExpectQuery("SELECT oi\\.id, oi\\.order_id, oi\\.item_id, COALESCE\\(i\\.name, ''\\), oi\\.quantity, oi\\.unit_price, oi\\.subtotal FROM order_items oi LEFT JOIN items i ON oi\\.item_id = i\\.id WHERE oi\\.order_id = \\$1").
					WithArgs(1).
					WillReturnRows(itemsRow)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var order models.Order
				err := json.Unmarshal(w.Body.Bytes(), &order)
				assert.NoError(t, err)
				assert.Equal(t, 1, order.ID)
				assert.Equal(t, "pending", order.Status)
				assert.Len(t, order.OrderItems, 1)
			},
		},
		{
			name:           "Returns 400 for invalid order ID",
			orderID:        "abc",
			setupMock:      func(m sqlmock.Sqlmock, id string) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid order ID", resp["error"])
			},
		},
		{
			name:    "Returns 404 when order not found",
			orderID: "999",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Order not found", resp["error"])
			},
		},
		{
			name:    "Returns 404 when database error occurs",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				m.ExpectQuery("SELECT o\\.id, o\\.customer_id, o\\.delivery_address, o\\.status, o\\.total_amount, o\\.notes, o\\.payment_method, o\\.scheduled_date, o\\.created_at, o\\.updated_at,").
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Order not found", resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock, tt.orderID)

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
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Creates order successfully",
			body: `{"customer_id":1,"delivery_address":"123 Main St","status":"pending","payment_method":"cash","items":[{"item_id":1,"quantity":2}]}`,
			setupMock: func(m sqlmock.Sqlmock, body string) {
				m.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM customers WHERE id = \\$1\\)").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

				m.ExpectBegin()
				m.ExpectQuery("SELECT price FROM items WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(12.99))
				m.ExpectQuery("INSERT INTO orders").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				m.ExpectQuery("SELECT price FROM items WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(12.99))
				m.ExpectExec("INSERT INTO order_items").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, float64(1), resp["id"])
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			body:           `{"customer_id":1,"items":[}`,
			setupMock:      func(m sqlmock.Sqlmock, body string) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp["error"], "invalid")
			},
		},
		{
			name: "Returns 400 for empty items array",
			body: `{"delivery_address":"123 Main St","items":[]}`,
			setupMock: func(m sqlmock.Sqlmock, body string) {
				m.ExpectBegin()
				m.ExpectRollback()
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "At least one item is required", resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock, tt.body)

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
		setupMock      func(sqlmock.Sqlmock, string, string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:    "Updates order successfully",
			orderID: "1",
			body:    `{"customer_id":1,"delivery_address":"123 Main St","status":"preparing","total_amount":30.00,"payment_method":"card","items":[{"item_id":1,"quantity":2}]}`,
			setupMock: func(m sqlmock.Sqlmock, id, body string) {
				m.ExpectBegin()
				m.ExpectExec("UPDATE orders SET customer_id = \\$1, delivery_address = \\$2, status = \\$3, total_amount = \\$4, notes = \\$5, payment_method = \\$6, scheduled_date = \\$7, updated_at = CURRENT_TIMESTAMP WHERE id = \\$8").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec("DELETE FROM order_items WHERE order_id = \\$1").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectQuery("SELECT price FROM items WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(12.99))
				m.ExpectExec("INSERT INTO order_items").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Order updated", resp["message"])
			},
		},
		{
			name:           "Returns 400 for invalid order ID",
			orderID:        "abc",
			body:           `{"status":"preparing"}`,
			setupMock:      func(m sqlmock.Sqlmock, id, body string) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid order ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			orderID:        "1",
			body:           `{"status":}`,
			setupMock:      func(m sqlmock.Sqlmock, id, body string) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "Returns 400 for invalid status",
			orderID:        "1",
			body:           `{"status":"invalid"}`,
			setupMock:      func(m sqlmock.Sqlmock, id, body string) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid status", resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock, tt.orderID, tt.body)

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
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:    "Deletes order successfully",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				m.ExpectExec("DELETE FROM orders WHERE id = \\$1").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Order deleted", resp["message"])
			},
		},
		{
			name:           "Returns 400 for invalid order ID",
			orderID:        "abc",
			setupMock:      func(m sqlmock.Sqlmock, id string) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid order ID", resp["error"])
			},
		},
		{
			name:    "Returns 500 on database error",
			orderID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				m.ExpectExec("DELETE FROM orders WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
		{
			name:    "Deletes non-existent order",
			orderID: "999",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				m.ExpectExec("DELETE FROM orders WHERE id = \\$1").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(999, 0))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Order deleted", resp["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
			defer db.Close()

			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock, tt.orderID)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.orderID}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/orders/"+tt.orderID, nil)

			DeleteOrder(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)

				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
