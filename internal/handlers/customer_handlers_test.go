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

// ── Column/query constants ───────────────────────────────────────────────────

var customerCols = []string{"id", "name", "phone", "email", "address", "created_at", "updated_at"}

const (
	customerQueryRegex       = `SELECT id, name, phone, email, address, created_at, updated_at\s+FROM customers`
	insertCustomerQueryRegex = `INSERT INTO customers`
	updateCustomerExecRegex  = `UPDATE customers\s+SET`
	orderCountQueryRegex     = `SELECT COUNT\(\*\) FROM orders WHERE customer_id = \$1`
	deleteCustomerExecRegex  = `DELETE FROM customers WHERE id = \$1`
)

// ── Row helpers ──────────────────────────────────────────────────────────────

// customerRow returns a sqlmock row representing a single customer.
func customerRow(id int, name, phone, email, address string) *sqlmock.Rows {
	return sqlmock.NewRows(customerCols).
		AddRow(id, name, phone, email, address, time.Now(), time.Now())
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestGetCustomers(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Returns all customers",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(customerQueryRegex).
					WillReturnRows(sqlmock.NewRows(customerCols).
						AddRow(1, "John Doe", "555-1234", "john@example.com", "123 Main St", time.Now(), time.Now()).
						AddRow(2, "Jane Smith", "555-5678", "jane@example.com", "456 Oak Ave", time.Now(), time.Now()))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customers []models.Customer
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &customers))
				assert.Len(t, customers, 2)
				assert.Equal(t, "John Doe", customers[0].Name)
				assert.Equal(t, "Jane Smith", customers[1].Name)
			},
		},
		{
			name: "Returns empty array when no customers exist",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(customerQueryRegex).
					WillReturnRows(sqlmock.NewRows(customerCols))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Should return null or [] — either is valid JSON for an empty slice
				assert.Contains(t, []string{"null", "[]"}, strings.TrimSpace(w.Body.String()))
			},
		},
		{
			name: "Returns 500 on database error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(customerQueryRegex).
					WillReturnError(fmt.Errorf("connection refused"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp["error"])
			},
		},
		{
			name: "Returns customers with all fields populated",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(customerQueryRegex).
					WillReturnRows(sqlmock.NewRows(customerCols).
						AddRow(1, "Alice", "555-0001", "alice@example.com", "1 A St",
							time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customers []models.Customer
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &customers))
				assert.Len(t, customers, 1)
				c := customers[0]
				assert.Equal(t, 1, c.ID)
				assert.Equal(t, "Alice", c.Name)
				assert.Equal(t, "555-0001", c.Phone)
				assert.Equal(t, "alice@example.com", c.Email)
				assert.Equal(t, "1 A St", c.Address)
			},
		},
		{
			name: "Continues scanning after a bad row",
			setupMock: func(m sqlmock.Sqlmock) {
				// Second row has a type mismatch (string for int ID) to force scan error.
				m.ExpectQuery(customerQueryRegex).
					WillReturnRows(sqlmock.NewRows(customerCols).
						AddRow("not-an-int", "Bad Row", "", "", "", time.Now(), time.Now()).
						AddRow(2, "Good Row", "555-0002", "good@example.com", "2 B St", time.Now(), time.Now()))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customers []models.Customer
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &customers))
				// Bad row skipped, good row present
				assert.Len(t, customers, 1)
				assert.Equal(t, "Good Row", customers[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodGet, "/api/customers", nil)

				GetCustomers(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

func TestGetCustomer(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Returns customer by ID",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(customerQueryRegex).
					WithArgs(1).
					WillReturnRows(customerRow(1, "John Doe", "555-1234", "john@example.com", "123 Main St"))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customer models.Customer
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &customer))
				assert.Equal(t, 1, customer.ID)
				assert.Equal(t, "John Doe", customer.Name)
				assert.Equal(t, "555-1234", customer.Phone)
				assert.Equal(t, "john@example.com", customer.Email)
				assert.Equal(t, "123 Main St", customer.Address)
			},
		},
		{
			name:           "Returns 400 for non-numeric ID",
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
			name:           "Returns 400 for zero ID",
			customerID:     "0",
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid customer ID", resp["error"])
			},
		},
		{
			name:       "Returns 404 when customer not found",
			customerID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(customerQueryRegex).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Customer not found", resp["error"])
			},
		},
		{
			name:       "Returns 404 on any database error",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(customerQueryRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("connection lost"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Customer not found", resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

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
		})
	}
}

func TestCreateCustomer(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Creates customer with all fields",
			body: `{"name":"John Doe","phone":"555-1234","email":"john@example.com","address":"123 Main St"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertCustomerQueryRegex).
					WithArgs("John Doe", "555-1234", "john@example.com", "123 Main St").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customer models.Customer
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &customer))
				assert.Equal(t, 1, customer.ID)
				assert.Equal(t, "John Doe", customer.Name)
				assert.Equal(t, "555-1234", customer.Phone)
				assert.Equal(t, "john@example.com", customer.Email)
				assert.Equal(t, "123 Main St", customer.Address)
			},
		},
		{
			name: "Creates customer with name only (optional fields empty)",
			body: `{"name":"Minimal Customer"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertCustomerQueryRegex).
					WithArgs("Minimal Customer", "", "", "").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var customer models.Customer
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &customer))
				assert.Equal(t, 2, customer.ID)
				assert.Equal(t, "Minimal Customer", customer.Name)
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			body:           `{invalid json`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp["error"])
			},
		},
		{
			name:           "Returns 400 when name is missing",
			body:           `{"phone":"555-1234","email":"test@example.com"}`,
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
			body:           `{"name":"   ","phone":"555-1234"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Name is required", resp["error"])
			},
		},
		{
			name: "Returns 500 on database error",
			body: `{"name":"John Doe","phone":"555-1234","email":"john@example.com","address":"123 Main St"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(insertCustomerQueryRegex).
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
				req := httptest.NewRequest(http.MethodPost, "/api/customers", strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				CreateCustomer(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

func TestUpdateCustomer(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		body           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Updates customer successfully",
			customerID: "1",
			body:       `{"name":"John Updated","phone":"555-9999","email":"updated@example.com","address":"789 New St"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateCustomerExecRegex).
					WithArgs("John Updated", "555-9999", "updated@example.com", "789 New St", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Customer updated", resp["message"])
			},
		},
		{
			name:       "Updates customer clearing optional fields",
			customerID: "2",
			body:       `{"name":"Name Only","phone":"","email":"","address":""}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateCustomerExecRegex).
					WithArgs("Name Only", "", "", "", 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Customer updated", resp["message"])
			},
		},
		{
			name:           "Returns 400 for non-numeric ID",
			customerID:     "abc",
			body:           `{"name":"John"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Invalid customer ID", resp["error"])
			},
		},
		{
			name:           "Returns 400 for invalid JSON",
			customerID:     "1",
			body:           `{invalid`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Returns 400 when name is missing",
			customerID:     "1",
			body:           `{"phone":"555-1234"}`,
			setupMock:      func(m sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Name is required", resp["error"])
			},
		},
		{
			name:       "Returns 500 on database error",
			customerID: "1",
			body:       `{"name":"John Doe"}`,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(updateCustomerExecRegex).
					WillReturnError(fmt.Errorf("deadlock detected"))
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
				c.Params = gin.Params{{Key: "id", Value: tt.customerID}}
				req := httptest.NewRequest(http.MethodPut, "/api/customers/"+tt.customerID, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				UpdateCustomer(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

func TestDeleteCustomer(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:       "Deletes customer with no orders",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderCountQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				m.ExpectExec(deleteCustomerExecRegex).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Customer deleted", resp["message"])
			},
		},
		{
			name:       "Returns 400 when customer has existing orders",
			customerID: "2",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderCountQueryRegex).
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				// No DELETE expected — handler should stop here
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Cannot delete customer with existing orders", resp["error"])
			},
		},
		{
			name:           "Returns 400 for non-numeric ID",
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
			name:       "Returns 500 when order count query fails",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderCountQueryRegex).
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
		{
			name:       "Returns 500 when delete query fails",
			customerID: "1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderCountQueryRegex).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				m.ExpectExec(deleteCustomerExecRegex).
					WithArgs(1).
					WillReturnError(fmt.Errorf("lock timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp["error"])
			},
		},
		{
			// Deleting a non-existent customer is idempotent — still returns 200
			name:       "Deletes non-existent customer without error",
			customerID: "999",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(orderCountQueryRegex).
					WithArgs(999).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				m.ExpectExec(deleteCustomerExecRegex).
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Customer deleted", resp["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{{Key: "id", Value: tt.customerID}}
				c.Request = httptest.NewRequest(http.MethodDelete, "/api/customers/"+tt.customerID, nil)

				DeleteCustomer(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}
