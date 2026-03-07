package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// setupTestDB creates a mock database for testing
func setupTestDB() (*sql.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	return db, mock, nil
}

// createTestContext creates a Gin context for testing
func createTestContext(db *sql.DB, method, path string, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Replace the global database.DB with our mock
	originalDB := database.DB
	database.DB = db

	// Restore after test
	database.DB = originalDB

	return c, w
}

// createCustomerJSON creates a JSON string for customer
func createCustomerJSON(name, phone, email, address string) string {
	customer := models.Customer{
		Name:    name,
		Phone:   phone,
		Email:   email,
		Address: address,
	}
	json, _ := json.Marshal(customer)
	return string(json)
}

// TestGetCustomers tests the GetCustomers handler
func TestGetCustomers(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedLen    int
	}{
		{
			name: "Returns empty list when no customers",
			setupMock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "email", "address", "created_at", "updated_at"})
				m.ExpectQuery("SELECT id, name, phone, email, address, created_at, updated_at FROM customers ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name: "Returns customers when they exist",
			setupMock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "email", "address", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "555-1234", "john@example.com", "123 Main St", time.Now(), time.Now()).
					AddRow(2, "Jane Doe", "555-5678", "jane@example.com", "456 Oak Ave", time.Now(), time.Now())
				m.ExpectQuery("SELECT id, name, phone, email, address, created_at, updated_at FROM customers ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name: "Returns 500 on database error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT id, name, phone, email, address, created_at, updated_at FROM customers ORDER BY created_at DESC").
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := setupTestDB()
			if err != nil {
				t.Fatalf("Failed to setup mock: %v", err)
			}
			defer db.Close()

			// Replace global database.DB with mock
			originalDB := database.DB
			database.DB = db
			defer func() { database.DB = originalDB }()

			tt.setupMock(mock)

			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/customers", nil)

			// Call handler
			GetCustomers(c)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// If expecting customers, verify them
			if tt.expectedStatus == http.StatusOK && tt.expectedLen > 0 {
				var customers []models.Customer
				err := json.Unmarshal(w.Body.Bytes(), &customers)
				assert.NoError(t, err)
				assert.Len(t, customers, tt.expectedLen)
			}

			// Verify all expectations
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestGetCustomer tests the GetCustomer handler
func TestGetCustomer(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Returns customer when found",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "email", "address", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "555-1234", "john@example.com", "123 Main St", time.Now(), time.Now())
				m.ExpectQuery("SELECT id, name, phone, email, address, created_at, updated_at FROM customers WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customer models.Customer
				err := json.Unmarshal(w.Body.Bytes(), &customer)
				assert.NoError(t, err)
				assert.Equal(t, "John Doe", customer.Name)
			},
		},
		{
			name:       "Returns 404 when customer not found",
			customerID: "999",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				m.ExpectQuery("SELECT id, name, phone, email, address, created_at, updated_at FROM customers WHERE id = \\$1").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse:  nil,
		},
		{
			name:       "Returns 400 for invalid customer ID",
			customerID: "abc",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				// No mock needed - should fail before DB call
			},
			expectedStatus: http.StatusBadRequest,
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

			if tt.setupMock != nil {
				tt.setupMock(mock, tt.customerID)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.customerID}}
			c.Request = httptest.NewRequest(http.MethodGet, "/api/customers/"+tt.customerID, nil)

			GetCustomer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestCreateCustomer tests the CreateCustomer handler
func TestCreateCustomer(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Creates customer successfully",
			body:           createCustomerJSON("John Doe", "555-1234", "john@example.com", "123 Main St"),
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customer models.Customer
				err := json.Unmarshal(w.Body.Bytes(), &customer)
				assert.NoError(t, err)
				assert.Equal(t, "John Doe", customer.Name)
				assert.Equal(t, "555-1234", customer.Phone)
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			body:           "{invalid json",
			expectedStatus: http.StatusBadRequest,
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

			if tt.name == "Creates customer successfully" {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("INSERT INTO customers").
					WillReturnRows(rows)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/customers", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			CreateCustomer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestUpdateCustomer tests the UpdateCustomer handler
func TestUpdateCustomer(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		body           string
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
	}{
		{
			name:           "Updates customer successfully",
			customerID:     "1",
			body:           createCustomerJSON("John Updated", "555-9999", "john.updated@example.com", "789 New St"),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Returns 400 for invalid customer ID",
			customerID:     "abc",
			body:           createCustomerJSON("John", "555-1234", "john@test.com", "123 St"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 for invalid JSON",
			customerID:     "1",
			body:           "{invalid",
			expectedStatus: http.StatusBadRequest,
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

			if tt.expectedStatus == http.StatusOK {
				mock.ExpectExec("UPDATE customers SET name = \\$1, phone = \\$2, email = \\$3, address = \\$4, updated_at = CURRENT_TIMESTAMP WHERE id = \\$5").
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.customerID}}
			c.Request = httptest.NewRequest(http.MethodPut, "/api/customers/"+tt.customerID, strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			UpdateCustomer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestDeleteCustomer tests the DeleteCustomer handler
func TestDeleteCustomer(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		setupMock      func(sqlmock.Sqlmock, string)
		expectedStatus int
		checkMessage   string
	}{
		{
			name:       "Deletes customer successfully",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				// First check for orders
				orderRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				m.ExpectQuery("SELECT COUNT\\(\\*\\) FROM orders WHERE customer_id = \\$1").
					WithArgs(1).
					WillReturnRows(orderRows)
				// Then delete
				m.ExpectExec("DELETE FROM customers WHERE id = \\$1").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkMessage:   "Customer deleted",
		},
		{
			name:       "Returns 400 when customer has orders",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				// Return error to simulate existing orders check failure
				m.ExpectQuery("SELECT COUNT\\(\\*\\) FROM orders WHERE customer_id = \\$1").
					WithArgs(1).
					WillReturnError(fmt.Errorf("foreign key constraint"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkMessage:   "",
		},
		{
			name:       "Returns 400 for invalid customer ID",
			customerID: "abc",
			setupMock: func(m sqlmock.Sqlmock, id string) {
				// No mock needed - fails before DB call
			},
			expectedStatus: http.StatusBadRequest,
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

			if tt.setupMock != nil {
				tt.setupMock(mock, tt.customerID)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.customerID}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/customers/"+tt.customerID, nil)

			DeleteCustomer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkMessage != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.checkMessage, response["message"])
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// BenchmarkGetCustomers benchmarks the GetCustomers handler
func BenchmarkGetCustomers(b *testing.B) {
	db, _, _ := setupTestDB()
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/customers", nil)
		GetCustomers(c)
	}
}
