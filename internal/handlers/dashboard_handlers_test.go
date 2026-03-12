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

func init() {
	gin.SetMode(gin.TestMode)
}

// ── Query regex constants ────────────────────────────────────────────────────

const (
	totalRevenueQueryRegex  = `SELECT COALESCE\(SUM\(total_amount\), 0\), COUNT\(\*\) FROM orders WHERE status != 'cancelled'`
	periodRevenueQueryRegex = `SELECT COALESCE\(SUM\(total_amount\), 0\), COUNT\(\*\) FROM orders WHERE status != 'cancelled' AND created_at >=`
	statusCountQueryRegex   = `SELECT status, COUNT\(\*\) FROM orders`
	bestSellingQueryRegex   = `SELECT i\.name, SUM\(oi\.quantity\)`
	topCustomersQueryRegex  = `SELECT COALESCE\(c\.name, 'Unknown'\)`
	salesTrendQueryRegex    = `SELECT DATE\(created_at AT TIME ZONE 'UTC'\)`
)

// ── Shared column slices ─────────────────────────────────────────────────────

var (
	revenueCountCols = []string{"coalesce", "count"}
	statusCols       = []string{"status", "count"}
	bestSellingCols  = []string{"name", "total_qty", "total_revenue"}
	topCustomerCols  = []string{"coalesce", "count", "sum"}
	salesTrendCols   = []string{"day", "orders", "revenue"}
)

// ── Mock helpers ─────────────────────────────────────────────────────────────

// mockRevenueSummary sets up the three revenue/count queries (total, monthly, daily).
// Pass UTC-based time values to match the refactored dashboard.go.
func mockRevenueSummary(m sqlmock.Sqlmock, total, monthly, daily float64, totalOrders, monthlyOrders, dailyOrders int) {
	now := time.Now().UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	m.ExpectQuery(totalRevenueQueryRegex).
		WillReturnRows(sqlmock.NewRows(revenueCountCols).AddRow(total, totalOrders))
	m.ExpectQuery(periodRevenueQueryRegex).
		WithArgs(startOfMonth).
		WillReturnRows(sqlmock.NewRows(revenueCountCols).AddRow(monthly, monthlyOrders))
	m.ExpectQuery(periodRevenueQueryRegex).
		WithArgs(startOfDay).
		WillReturnRows(sqlmock.NewRows(revenueCountCols).AddRow(daily, dailyOrders))
}

// mockStatusCounts sets up the orders-by-status query.
func mockStatusCounts(m sqlmock.Sqlmock, rows *sqlmock.Rows) {
	m.ExpectQuery(statusCountQueryRegex).WillReturnRows(rows)
}

// mockBestSelling sets up the best selling items query.
func mockBestSelling(m sqlmock.Sqlmock, rows *sqlmock.Rows) {
	m.ExpectQuery(bestSellingQueryRegex).WillReturnRows(rows)
}

// mockTopCustomers sets up the top customers query.
func mockTopCustomers(m sqlmock.Sqlmock, rows *sqlmock.Rows) {
	m.ExpectQuery(topCustomersQueryRegex).WillReturnRows(rows)
}

// mockSalesTrend sets up the single GROUP BY sales trend query.
// The refactored code fires one query (not 30), returning aggregated rows.
func mockSalesTrend(m sqlmock.Sqlmock, rows *sqlmock.Rows) {
	m.ExpectQuery(salesTrendQueryRegex).WillReturnRows(rows)
}

// mockEmptySalesTrend sets up the sales trend query returning no rows.
func mockEmptySalesTrend(m sqlmock.Sqlmock) {
	mockSalesTrend(m, sqlmock.NewRows(salesTrendCols))
}

// mockFullDashboard sets up all queries for a standard successful response.
func mockFullDashboard(m sqlmock.Sqlmock) {
	mockRevenueSummary(m, 1500.00, 500.00, 100.00, 10, 5, 2)

	mockStatusCounts(m, sqlmock.NewRows(statusCols).
		AddRow("pending", 3).
		AddRow("preparing", 2).
		AddRow("ready", 1).
		AddRow("delivered", 4))

	mockBestSelling(m, sqlmock.NewRows(bestSellingCols).
		AddRow("Pizza", 50, 500.00).
		AddRow("Burger", 30, 300.00))

	mockTopCustomers(m, sqlmock.NewRows(topCustomerCols).
		AddRow("John Doe", 5, 500.00).
		AddRow("Jane Smith", 3, 300.00))

	now := time.Now().UTC()
	trendRows := sqlmock.NewRows(salesTrendCols)
	for i := salesTrendDays - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		trendRows.AddRow(date, 1, 50.00)
	}
	mockSalesTrend(m, trendRows)
}

// runDashboardRequest fires a GET /api/dashboard request and returns the recorder.
func runDashboardRequest(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	GetDashboardStats(c)
	return w
}

// withMockDB sets database.DB to a mock, runs f, then restores the original.
func withMockDB(t *testing.T, f func(m sqlmock.Sqlmock)) {
	t.Helper()
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	f(mock)
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestGetDashboardStats(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Returns full dashboard stats successfully",
			setupMock:      mockFullDashboard,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
				assert.Equal(t, 1500.00, stats.TotalRevenue)
				assert.Equal(t, 10, stats.TotalOrders)
				assert.Equal(t, 500.00, stats.MonthlyRevenue)
				assert.Equal(t, 5, stats.MonthlyOrders)
				assert.Equal(t, 100.00, stats.DailyRevenue)
				assert.Equal(t, 2, stats.DailyOrders)
				assert.Equal(t, 150.00, stats.AverageOrderValue)
				assert.Len(t, stats.OrdersByStatus, 4)
				assert.Len(t, stats.BestSellingItems, 2)
				assert.Len(t, stats.TopCustomers, 2)
				assert.Len(t, stats.SalesTrend, salesTrendDays)
			},
		},
		{
			name: "Returns zeros and empty slices when no orders",
			setupMock: func(m sqlmock.Sqlmock) {
				mockRevenueSummary(m, 0, 0, 0, 0, 0, 0)
				mockStatusCounts(m, sqlmock.NewRows(statusCols))
				mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
				mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
				mockEmptySalesTrend(m)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
				assert.Equal(t, 0.00, stats.TotalRevenue)
				assert.Equal(t, 0, stats.TotalOrders)
				assert.Equal(t, 0.00, stats.AverageOrderValue)
				assert.Empty(t, stats.OrdersByStatus)
				assert.Empty(t, stats.BestSellingItems)
				assert.Empty(t, stats.TopCustomers)
				// SalesTrend still has salesTrendDays entries — all zeros
				assert.Len(t, stats.SalesTrend, salesTrendDays)
			},
		},
		{
			name: "Returns 500 when revenue summary query fails",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(totalRevenueQueryRegex).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Partial data returned when supplementary queries fail",
			setupMock: func(m sqlmock.Sqlmock) {
				mockRevenueSummary(m, 500.00, 100.00, 50.00, 5, 2, 1)
				// Status, items, customers, trend queries all fail —
				// dashboard still returns 200 with revenue data populated.
				m.ExpectQuery(statusCountQueryRegex).WillReturnError(fmt.Errorf("db error"))
				m.ExpectQuery(bestSellingQueryRegex).WillReturnError(fmt.Errorf("db error"))
				m.ExpectQuery(topCustomersQueryRegex).WillReturnError(fmt.Errorf("db error"))
				m.ExpectQuery(salesTrendQueryRegex).WillReturnError(fmt.Errorf("db error"))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
				// Revenue data is present
				assert.Equal(t, 500.00, stats.TotalRevenue)
				assert.Equal(t, 5, stats.TotalOrders)
				// Supplementary data is empty but not an error
				assert.Empty(t, stats.OrdersByStatus)
				assert.Empty(t, stats.BestSellingItems)
				assert.Empty(t, stats.TopCustomers)
				assert.Empty(t, stats.SalesTrend)
			},
		},
		{
			name: "Average order value is zero when total orders is zero",
			setupMock: func(m sqlmock.Sqlmock) {
				mockRevenueSummary(m, 0, 0, 0, 0, 0, 0)
				mockStatusCounts(m, sqlmock.NewRows(statusCols))
				mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
				mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
				mockEmptySalesTrend(m)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
				// Guard against divide-by-zero
				assert.Equal(t, 0.00, stats.AverageOrderValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)
				w := runDashboardRequest(t)
				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

func TestDashboardBestSellingItems(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		mockRevenueSummary(m, 1000.00, 500.00, 100.00, 10, 5, 1)
		mockStatusCounts(m, sqlmock.NewRows(statusCols).AddRow("delivered", 10))
		mockBestSelling(m, sqlmock.NewRows(bestSellingCols).
			AddRow("Pepperoni Pizza", 25, 375.00).
			AddRow("Cheese Burger", 20, 300.00).
			AddRow("Caesar Salad", 15, 150.00))
		mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
		mockEmptySalesTrend(m)

		w := runDashboardRequest(t)
		assert.Equal(t, http.StatusOK, w.Code)

		var stats DashboardStats
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
		assert.Len(t, stats.BestSellingItems, 3)
		assert.Equal(t, "Pepperoni Pizza", stats.BestSellingItems[0].Name)
		assert.Equal(t, 25, stats.BestSellingItems[0].Quantity)
		assert.Equal(t, 375.00, stats.BestSellingItems[0].Revenue)
		assert.Equal(t, "Caesar Salad", stats.BestSellingItems[2].Name)

		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestDashboardTopCustomers(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		mockRevenueSummary(m, 1000.00, 500.00, 100.00, 10, 5, 1)
		mockStatusCounts(m, sqlmock.NewRows(statusCols).AddRow("delivered", 10))
		mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
		mockTopCustomers(m, sqlmock.NewRows(topCustomerCols).
			AddRow("Alice Johnson", 8, 800.00).
			AddRow("Bob Smith", 5, 500.00).
			AddRow("Carol Williams", 3, 300.00).
			AddRow("David Brown", 2, 200.00).
			AddRow("Eve Davis", 1, 100.00))
		mockEmptySalesTrend(m)

		w := runDashboardRequest(t)
		assert.Equal(t, http.StatusOK, w.Code)

		var stats DashboardStats
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
		assert.Len(t, stats.TopCustomers, 5)
		assert.Equal(t, "Alice Johnson", stats.TopCustomers[0].Name)
		assert.Equal(t, 8, stats.TopCustomers[0].OrderCount)
		assert.Equal(t, 800.00, stats.TopCustomers[0].TotalSpent)
		assert.Equal(t, "Bob Smith", stats.TopCustomers[1].Name)

		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestDashboardSalesTrend(t *testing.T) {
	t.Run("Trend covers exactly salesTrendDays days", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			mockRevenueSummary(m, 0, 0, 0, 0, 0, 0)
			mockStatusCounts(m, sqlmock.NewRows(statusCols))
			mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
			mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))

			// Return data for only a subset of days — the rest should be zero-filled.
			now := time.Now().UTC()
			trendRows := sqlmock.NewRows(salesTrendCols).
				AddRow(now.AddDate(0, 0, -1).Format("2006-01-02"), 3, 150.00).
				AddRow(now.Format("2006-01-02"), 5, 250.00)
			mockSalesTrend(m, trendRows)

			w := runDashboardRequest(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var stats DashboardStats
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
			assert.Len(t, stats.SalesTrend, salesTrendDays)

			// Last two days should have data; the rest should be zeros.
			last := stats.SalesTrend[salesTrendDays-1]
			assert.Equal(t, now.Format("2006-01-02"), last.Date)
			assert.Equal(t, 5, last.Orders)
			assert.Equal(t, 250.00, last.Revenue)

			secondLast := stats.SalesTrend[salesTrendDays-2]
			assert.Equal(t, 3, secondLast.Orders)

			// A day with no orders should be zero-filled, not absent.
			first := stats.SalesTrend[0]
			assert.Equal(t, 0, first.Orders)
			assert.Equal(t, 0.00, first.Revenue)

			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Trend dates are in ascending order", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			mockRevenueSummary(m, 0, 0, 0, 0, 0, 0)
			mockStatusCounts(m, sqlmock.NewRows(statusCols))
			mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
			mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
			mockEmptySalesTrend(m)

			w := runDashboardRequest(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var stats DashboardStats
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
			assert.Len(t, stats.SalesTrend, salesTrendDays)

			for i := 1; i < len(stats.SalesTrend); i++ {
				assert.True(t, stats.SalesTrend[i].Date > stats.SalesTrend[i-1].Date,
					"expected ascending dates at index %d: %s <= %s",
					i, stats.SalesTrend[i-1].Date, stats.SalesTrend[i].Date)
			}

			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestDashboardExcludesCancelledOrders(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		// Mock returns only non-cancelled counts — verifies the SQL WHERE clause
		// is respected by checking the response reflects those values exactly.
		mockRevenueSummary(m, 100.00, 100.00, 100.00, 2, 2, 2)
		mockStatusCounts(m, sqlmock.NewRows(statusCols).AddRow("delivered", 2))
		mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
		mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
		mockEmptySalesTrend(m)

		w := runDashboardRequest(t)
		assert.Equal(t, http.StatusOK, w.Code)

		var stats DashboardStats
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
		assert.Equal(t, 2, stats.TotalOrders)
		assert.Equal(t, 100.00, stats.TotalRevenue)
		assert.Equal(t, map[string]int{"delivered": 2}, stats.OrdersByStatus)

		assert.NoError(t, m.ExpectationsWereMet())
	})
}
