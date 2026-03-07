package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"food-order-tracking/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDashboardStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Returns dashboard stats successfully",
			setupMock: func(m sqlmock.Sqlmock) {
				now := time.Now()
				startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

				totalRow := sqlmock.NewRows([]string{"coalesce", "count"}).
					AddRow(1500.00, 10)
				m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled'").
					WillReturnRows(totalRow)

				monthlyRow := sqlmock.NewRows([]string{"coalesce", "count"}).
					AddRow(500.00, 5)
				m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").
					WithArgs(startOfMonth).
					WillReturnRows(monthlyRow)

				dailyRow := sqlmock.NewRows([]string{"coalesce", "count"}).
					AddRow(100.00, 2)
				m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").
					WithArgs(startOfDay).
					WillReturnRows(dailyRow)

				statusRow := sqlmock.NewRows([]string{"status", "count"}).
					AddRow("pending", 3).
					AddRow("preparing", 2).
					AddRow("ready", 1).
					AddRow("delivered", 4)
				m.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM orders WHERE status IS NOT NULL AND status != '' GROUP BY status").
					WillReturnRows(statusRow)

				itemRow := sqlmock.NewRows([]string{"name", "total_qty", "total_revenue"}).
					AddRow("Pizza", 50, 500.00).
					AddRow("Burger", 30, 300.00)
				m.ExpectQuery("SELECT i.name, SUM\\(oi.quantity\\) as total_qty, SUM\\(oi.subtotal\\) as total_revenue FROM order_items oi JOIN items i ON oi.item_id = i.id JOIN orders o ON oi.order_id = o.id WHERE o.status != 'cancelled' GROUP BY i.id, i.name ORDER BY total_qty DESC LIMIT 10").
					WillReturnRows(itemRow)

				customerRow := sqlmock.NewRows([]string{"coalesce", "count", "sum"}).
					AddRow("John Doe", 5, 500.00).
					AddRow("Jane Smith", 3, 300.00)
				m.ExpectQuery("SELECT COALESCE\\(c.name, 'Unknown'\\), COUNT\\(\\*\\), SUM\\(o.total_amount\\) FROM orders o LEFT JOIN customers c ON o.customer_id = c.id WHERE o.status != 'cancelled' GROUP BY o.customer_id, c.name ORDER BY COUNT\\(\\*\\) DESC LIMIT 5").
					WillReturnRows(customerRow)

				for i := 29; i >= 0; i-- {
					date := now.AddDate(0, 0, -i)
					startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
					endOfDate := startOfDate.AddDate(0, 0, 1)

					trendRow := sqlmock.NewRows([]string{"coalesce", "count"}).
						AddRow(50.00, 1)
					m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >= .* AND created_at <").
						WithArgs(startOfDate, endOfDate).
						WillReturnRows(trendRow)
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				err := json.Unmarshal(w.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Equal(t, float64(1500), stats.TotalRevenue)
				assert.Equal(t, 10, stats.TotalOrders)
				assert.Equal(t, float64(500), stats.MonthlyRevenue)
				assert.Equal(t, 5, stats.MonthlyOrders)
				assert.Equal(t, float64(100), stats.DailyRevenue)
				assert.Equal(t, 2, stats.DailyOrders)
				assert.Equal(t, 150.0, stats.AverageOrderValue)
				assert.Len(t, stats.OrdersByStatus, 4)
				assert.Len(t, stats.BestSellingItems, 2)
				assert.Len(t, stats.TopCustomers, 2)
				assert.Len(t, stats.SalesTrend, 30)
			},
		},
		{
			name: "Returns zeros when no orders",
			setupMock: func(m sqlmock.Sqlmock) {
				now := time.Now()
				startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

				totalRow := sqlmock.NewRows([]string{"coalesce", "count"}).
					AddRow(0, 0)
				m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled'").
					WillReturnRows(totalRow)

				monthlyRow := sqlmock.NewRows([]string{"coalesce", "count"}).
					AddRow(0, 0)
				m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").
					WithArgs(startOfMonth).
					WillReturnRows(monthlyRow)

				dailyRow := sqlmock.NewRows([]string{"coalesce", "count"}).
					AddRow(0, 0)
				m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").
					WithArgs(startOfDay).
					WillReturnRows(dailyRow)

				statusRow := sqlmock.NewRows([]string{"status", "count"})
				m.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM orders WHERE status IS NOT NULL AND status != '' GROUP BY status").
					WillReturnRows(statusRow)

				itemRow := sqlmock.NewRows([]string{"name", "total_qty", "total_revenue"})
				m.ExpectQuery("SELECT i.name, SUM\\(oi.quantity\\) as total_qty, SUM\\(oi.subtotal\\) as total_revenue FROM order_items oi JOIN items i ON oi.item_id = i.id JOIN orders o ON oi.order_id = o.id WHERE o.status != 'cancelled' GROUP BY i.id, i.name ORDER BY total_qty DESC LIMIT 10").
					WillReturnRows(itemRow)

				customerRow := sqlmock.NewRows([]string{"coalesce", "count", "sum"})
				m.ExpectQuery("SELECT COALESCE\\(c.name, 'Unknown'\\), COUNT\\(\\*\\), SUM\\(o.total_amount\\) FROM orders o LEFT JOIN customers c ON o.customer_id = c.id WHERE o.status != 'cancelled' GROUP BY o.customer_id, c.name ORDER BY COUNT\\(\\*\\) DESC LIMIT 5").
					WillReturnRows(customerRow)

				for i := 29; i >= 0; i-- {
					date := now.AddDate(0, 0, -i)
					startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
					endOfDate := startOfDate.AddDate(0, 0, 1)

					trendRow := sqlmock.NewRows([]string{"coalesce", "count"}).
						AddRow(0, 0)
					m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >= .* AND created_at <").
						WithArgs(startOfDate, endOfDate).
						WillReturnRows(trendRow)
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				err := json.Unmarshal(w.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Equal(t, float64(0), stats.TotalRevenue)
				assert.Equal(t, 0, stats.TotalOrders)
				assert.Equal(t, 0.0, stats.AverageOrderValue)
			},
		},
		{
			name: "Returns 500 on database error",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled'").
					WillReturnError(fmt.Errorf("database connection error"))
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

			tt.setupMock(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)

			GetDashboardStats(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDashboardStatsBestSellingItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup mock: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	totalRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(1000.00, 10)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled'").WillReturnRows(totalRow)

	monthlyRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(500.00, 5)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").WithArgs(startOfMonth).WillReturnRows(monthlyRow)

	dailyRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(100.00, 1)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").WithArgs(startOfDay).WillReturnRows(dailyRow)

	statusRow := sqlmock.NewRows([]string{"status", "count"}).AddRow("delivered", 10)
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM orders WHERE status IS NOT NULL AND status != '' GROUP BY status").WillReturnRows(statusRow)

	itemRow := sqlmock.NewRows([]string{"name", "total_qty", "total_revenue"}).
		AddRow("Pepperoni Pizza", 25, 375.00).
		AddRow("Cheese Burger", 20, 300.00).
		AddRow("Caesar Salad", 15, 150.00)
	mock.ExpectQuery("SELECT i.name, SUM\\(oi.quantity\\) as total_qty, SUM\\(oi.subtotal\\) as total_revenue FROM order_items oi JOIN items i ON oi.item_id = i.id JOIN orders o ON oi.order_id = o.id WHERE o.status != 'cancelled' GROUP BY i.id, i.name ORDER BY total_qty DESC LIMIT 10").WillReturnRows(itemRow)

	customerRow := sqlmock.NewRows([]string{"coalesce", "count", "sum"})
	mock.ExpectQuery("SELECT COALESCE\\(c.name, 'Unknown'\\), COUNT\\(\\*\\), SUM\\(o.total_amount\\) FROM orders o LEFT JOIN customers c ON o.customer_id = c.id WHERE o.status != 'cancelled' GROUP BY o.customer_id, c.name ORDER BY COUNT\\(\\*\\) DESC LIMIT 5").WillReturnRows(customerRow)

	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDate := startOfDate.AddDate(0, 0, 1)

		trendRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(33.33, 1)
		mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >= .* AND created_at <").WithArgs(startOfDate, endOfDate).WillReturnRows(trendRow)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)

	GetDashboardStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats DashboardStats
	err = json.Unmarshal(w.Body.Bytes(), &stats)
	assert.NoError(t, err)
	assert.Len(t, stats.BestSellingItems, 3)
	assert.Equal(t, "Pepperoni Pizza", stats.BestSellingItems[0].Name)
	assert.Equal(t, 25, stats.BestSellingItems[0].Quantity)
	assert.Equal(t, 375.00, stats.BestSellingItems[0].Revenue)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardStatsTopCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup mock: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	totalRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(1000.00, 10)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled'").WillReturnRows(totalRow)

	monthlyRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(500.00, 5)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").WithArgs(startOfMonth).WillReturnRows(monthlyRow)

	dailyRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(100.00, 1)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").WithArgs(startOfDay).WillReturnRows(dailyRow)

	statusRow := sqlmock.NewRows([]string{"status", "count"}).AddRow("delivered", 10)
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM orders WHERE status IS NOT NULL AND status != '' GROUP BY status").WillReturnRows(statusRow)

	itemRow := sqlmock.NewRows([]string{"name", "total_qty", "total_revenue"})
	mock.ExpectQuery("SELECT i.name, SUM\\(oi.quantity\\) as total_qty, SUM\\(oi.subtotal\\) as total_revenue FROM order_items oi JOIN items i ON oi.item_id = i.id JOIN orders o ON oi.order_id = o.id WHERE o.status != 'cancelled' GROUP BY i.id, i.name ORDER BY total_qty DESC LIMIT 10").WillReturnRows(itemRow)

	customerRow := sqlmock.NewRows([]string{"coalesce", "count", "sum"}).
		AddRow("Alice Johnson", 8, 800.00).
		AddRow("Bob Smith", 5, 500.00).
		AddRow("Carol Williams", 3, 300.00).
		AddRow("David Brown", 2, 200.00).
		AddRow("Eve Davis", 1, 100.00)
	mock.ExpectQuery("SELECT COALESCE\\(c.name, 'Unknown'\\), COUNT\\(\\*\\), SUM\\(o.total_amount\\) FROM orders o LEFT JOIN customers c ON o.customer_id = c.id WHERE o.status != 'cancelled' GROUP BY o.customer_id, c.name ORDER BY COUNT\\(\\*\\) DESC LIMIT 5").WillReturnRows(customerRow)

	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDate := startOfDate.AddDate(0, 0, 1)

		trendRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(33.33, 1)
		mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >= .* AND created_at <").WithArgs(startOfDate, endOfDate).WillReturnRows(trendRow)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)

	GetDashboardStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats DashboardStats
	err = json.Unmarshal(w.Body.Bytes(), &stats)
	assert.NoError(t, err)
	assert.Len(t, stats.TopCustomers, 5)
	assert.Equal(t, "Alice Johnson", stats.TopCustomers[0].Name)
	assert.Equal(t, 8, stats.TopCustomers[0].OrderCount)
	assert.Equal(t, 800.00, stats.TopCustomers[0].TotalSpent)

	assert.Equal(t, "Bob Smith", stats.TopCustomers[1].Name)
	assert.Equal(t, 5, stats.TopCustomers[1].OrderCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardStatsExcludeCancelled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup mock: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	totalRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(100.00, 2)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled'").WillReturnRows(totalRow)

	monthlyRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(100.00, 2)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").WithArgs(startOfMonth).WillReturnRows(monthlyRow)

	dailyRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(100.00, 2)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >=").WithArgs(startOfDay).WillReturnRows(dailyRow)

	statusRow := sqlmock.NewRows([]string{"status", "count"}).AddRow("delivered", 2)
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM orders WHERE status IS NOT NULL AND status != '' GROUP BY status").WillReturnRows(statusRow)

	itemRow := sqlmock.NewRows([]string{"name", "total_qty", "total_revenue"})
	mock.ExpectQuery("SELECT i.name, SUM\\(oi.quantity\\) as total_qty, SUM\\(oi.subtotal\\) as total_revenue FROM order_items oi JOIN items i ON oi.item_id = i.id JOIN orders o ON oi.order_id = o.id WHERE o.status != 'cancelled' GROUP BY i.id, i.name ORDER BY total_qty DESC LIMIT 10").WillReturnRows(itemRow)

	customerRow := sqlmock.NewRows([]string{"coalesce", "count", "sum"})
	mock.ExpectQuery("SELECT COALESCE\\(c.name, 'Unknown'\\), COUNT\\(\\*\\), SUM\\(o.total_amount\\) FROM orders o LEFT JOIN customers c ON o.customer_id = c.id WHERE o.status != 'cancelled' GROUP BY o.customer_id, c.name ORDER BY COUNT\\(\\*\\) DESC LIMIT 5").WillReturnRows(customerRow)

	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDate := startOfDate.AddDate(0, 0, 1)

		trendRow := sqlmock.NewRows([]string{"coalesce", "count"}).AddRow(0, 0)
		mock.ExpectQuery("SELECT COALESCE\\(SUM\\(total_amount\\), 0\\), COUNT\\(\\*\\) FROM orders WHERE status != 'cancelled' AND created_at >= .* AND created_at <").WithArgs(startOfDate, endOfDate).WillReturnRows(trendRow)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)

	GetDashboardStats(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
