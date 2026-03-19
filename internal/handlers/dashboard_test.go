package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ── Query regex constants ────────────────────────────────────────────────────

const (
	totalRevenueQueryRegex       = `SELECT COALESCE\(SUM\(total_amount\), 0\), COUNT\(\*\) FROM orders WHERE status != 'cancelled'`
	periodRevenueQueryRegex      = `SELECT COALESCE\(SUM\(total_amount\), 0\), COUNT\(\*\) FROM orders WHERE status != 'cancelled' AND created_at >=`
	statusCountQueryRegex        = `SELECT status, COUNT\(\*\) FROM orders`
	bestSellingQueryRegex        = `SELECT i\.name, SUM\(oi\.quantity\)`
	topCustomersQueryRegex       = `SELECT COALESCE\(c\.name, 'Unknown'\)`
	salesTrendQueryRegex         = `SELECT DATE\(created_at AT TIME ZONE 'UTC'\)`
	modifierPopularityQueryRegex = `WITH modifier_counts AS`
)

// ── Shared column slices ─────────────────────────────────────────────────────

var (
	revenueCountCols       = []string{"coalesce", "count"}
	statusCols             = []string{"status", "count"}
	bestSellingCols        = []string{"name", "total_qty", "total_revenue"}
	topCustomerCols        = []string{"coalesce", "count", "sum"}
	salesTrendCols         = []string{"day", "orders", "revenue"}
	modifierPopularityCols = []string{
		"item_id", "item_name", "modifier_name", "price_adjustment",
		"times_ordered", "pct_of_orders", "revenue",
		"total_item_orders", "top_customer",
	}
)

// ── Mock helpers ─────────────────────────────────────────────────────────────

// mockRevenueSummary sets up the three revenue/count queries (total, monthly, daily).
func mockRevenueSummary(m sqlmock.Sqlmock, total, monthly, daily float64, totalOrders, monthlyOrders, dailyOrders int) {
	m.ExpectQuery(totalRevenueQueryRegex).
		WillReturnRows(sqlmock.NewRows(revenueCountCols).AddRow(total, totalOrders))
	m.ExpectQuery(periodRevenueQueryRegex).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(revenueCountCols).AddRow(monthly, monthlyOrders))
	m.ExpectQuery(periodRevenueQueryRegex).
		WithArgs(sqlmock.AnyArg()).
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
func mockSalesTrend(m sqlmock.Sqlmock, rows *sqlmock.Rows) {
	m.ExpectQuery(salesTrendQueryRegex).WithArgs(sqlmock.AnyArg()).WillReturnRows(rows)
}

// mockEmptySalesTrend sets up the sales trend query returning no rows.
func mockEmptySalesTrend(m sqlmock.Sqlmock) {
	mockSalesTrend(m, sqlmock.NewRows(salesTrendCols))
}

// mockModifierPopularity sets up the modifier popularity CTE query.
func mockModifierPopularity(m sqlmock.Sqlmock, rows *sqlmock.Rows) {
	m.ExpectQuery(modifierPopularityQueryRegex).WillReturnRows(rows)
}

// mockEmptyModifierPopularity sets up the modifier popularity query returning no rows.
func mockEmptyModifierPopularity(m sqlmock.Sqlmock) {
	mockModifierPopularity(m, sqlmock.NewRows(modifierPopularityCols))
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

	trendNow := time.Now().UTC()
	trendRows := sqlmock.NewRows(salesTrendCols)
	for i := salesTrendDays - 1; i >= 0; i-- {
		date := trendNow.AddDate(0, 0, -i).Format("2006-01-02")
		trendRows.AddRow(date, 1, 50.00)
	}
	mockSalesTrend(m, trendRows)

	mockModifierPopularity(m, sqlmock.NewRows(modifierPopularityCols).
		AddRow(1, "Pizza", "Extra Cheese", 1.50, 30, 60.0, 45.00, 50, "John Doe").
		AddRow(1, "Pizza", "Mushrooms", 0.75, 15, 30.0, 11.25, 50, "Jane Smith").
		AddRow(2, "Burger", "Bacon", 2.00, 20, 66.7, 40.00, 30, "John Doe"))
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
				assert.Equal(t, 1500.00/10, stats.AverageOrderValue)
				assert.Len(t, stats.OrdersByStatus, 4)
				assert.Len(t, stats.BestSellingItems, 2)
				assert.Len(t, stats.TopCustomers, 2)
				assert.Len(t, stats.SalesTrend, salesTrendDays)
				assert.Len(t, stats.ModifierPopularity, 2)
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
				mockEmptyModifierPopularity(m)
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
				assert.Len(t, stats.SalesTrend, salesTrendDays)
				assert.Empty(t, stats.ModifierPopularity)
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
				m.ExpectQuery(statusCountQueryRegex).WillReturnError(fmt.Errorf("db error"))
				m.ExpectQuery(bestSellingQueryRegex).WillReturnError(fmt.Errorf("db error"))
				m.ExpectQuery(topCustomersQueryRegex).WillReturnError(fmt.Errorf("db error"))
				m.ExpectQuery(salesTrendQueryRegex).WithArgs(sqlmock.AnyArg()).WillReturnError(fmt.Errorf("db error"))
				m.ExpectQuery(modifierPopularityQueryRegex).WillReturnError(fmt.Errorf("db error"))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
				assert.Equal(t, 500.00, stats.TotalRevenue)
				assert.Equal(t, 5, stats.TotalOrders)
				assert.Empty(t, stats.OrdersByStatus)
				assert.Empty(t, stats.BestSellingItems)
				assert.Empty(t, stats.TopCustomers)
				assert.Empty(t, stats.SalesTrend)
				assert.Empty(t, stats.ModifierPopularity)
				assert.Len(t, stats.Warnings, 5)
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
				mockEmptyModifierPopularity(m)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var stats DashboardStats
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
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
		mockEmptyModifierPopularity(m)

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
		mockEmptyModifierPopularity(m)

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

			now := time.Now().UTC()
			trendRows := sqlmock.NewRows(salesTrendCols).
				AddRow(now.AddDate(0, 0, -1).Format("2006-01-02"), 3, 150.00).
				AddRow(now.Format("2006-01-02"), 5, 250.00)
			mockSalesTrend(m, trendRows)
			mockEmptyModifierPopularity(m)

			w := runDashboardRequest(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var stats DashboardStats
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
			assert.Len(t, stats.SalesTrend, salesTrendDays)

			last := stats.SalesTrend[salesTrendDays-1]
			assert.Equal(t, now.Format("2006-01-02"), last.Date)
			assert.Equal(t, 5, last.Orders)
			assert.Equal(t, 250.00, last.Revenue)

			secondLast := stats.SalesTrend[salesTrendDays-2]
			assert.Equal(t, 3, secondLast.Orders)

			emptyDay := stats.SalesTrend[salesTrendDays-3]
			assert.Equal(t, 0, emptyDay.Orders)
			assert.Equal(t, 0.00, emptyDay.Revenue)

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
			mockEmptyModifierPopularity(m)

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
		mockRevenueSummary(m, 100.00, 100.00, 100.00, 2, 2, 2)
		mockStatusCounts(m, sqlmock.NewRows(statusCols).AddRow("delivered", 2))
		mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
		mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
		mockEmptySalesTrend(m)
		mockEmptyModifierPopularity(m)

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

func TestDashboardModifierPopularity(t *testing.T) {
	t.Run("Groups modifiers correctly by item", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			mockRevenueSummary(m, 1000.00, 500.00, 100.00, 10, 5, 1)
			mockStatusCounts(m, sqlmock.NewRows(statusCols).AddRow("delivered", 10))
			mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
			mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
			mockEmptySalesTrend(m)
			mockModifierPopularity(m, sqlmock.NewRows(modifierPopularityCols).
				AddRow(1, "Pizza", "Extra Cheese", 1.50, 30, 60.0, 45.00, 50, "John Doe").
				AddRow(1, "Pizza", "Mushrooms", 0.75, 15, 30.0, 11.25, 50, "Jane Smith").
				AddRow(2, "Burger", "Bacon", 2.00, 20, 66.7, 40.00, 30, "John Doe"))

			w := runDashboardRequest(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var stats DashboardStats
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
			assert.Len(t, stats.ModifierPopularity, 2)

			pizza := stats.ModifierPopularity[0]
			assert.Equal(t, "Pizza", pizza.ItemName)
			assert.Equal(t, 50, pizza.TotalItemOrders)
			assert.Len(t, pizza.Modifiers, 2)
			assert.Equal(t, "Extra Cheese", pizza.Modifiers[0].ModifierName)
			assert.Equal(t, 30, pizza.Modifiers[0].TimesOrdered)
			assert.Equal(t, 60.0, pizza.Modifiers[0].PctOfOrders)
			assert.Equal(t, 45.00, pizza.Modifiers[0].Revenue)
			assert.Equal(t, "John Doe", pizza.Modifiers[0].TopCustomer)
			assert.Equal(t, "Mushrooms", pizza.Modifiers[1].ModifierName)

			burger := stats.ModifierPopularity[1]
			assert.Equal(t, "Burger", burger.ItemName)
			assert.Len(t, burger.Modifiers, 1)
			assert.Equal(t, "Bacon", burger.Modifiers[0].ModifierName)
			assert.Equal(t, 2.00, burger.Modifiers[0].PriceAdjustment)

			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Empty when no modifiers have been ordered", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			mockRevenueSummary(m, 0, 0, 0, 0, 0, 0)
			mockStatusCounts(m, sqlmock.NewRows(statusCols))
			mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
			mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
			mockEmptySalesTrend(m)
			mockEmptyModifierPopularity(m)

			w := runDashboardRequest(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var stats DashboardStats
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
			assert.Empty(t, stats.ModifierPopularity)

			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Modifiers within item are ordered by times_ordered descending", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			mockRevenueSummary(m, 0, 0, 0, 0, 0, 0)
			mockStatusCounts(m, sqlmock.NewRows(statusCols))
			mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
			mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
			mockEmptySalesTrend(m)
			// DB returns rows already sorted by times_ordered DESC (enforced by ORDER BY in SQL).
			mockModifierPopularity(m, sqlmock.NewRows(modifierPopularityCols).
				AddRow(1, "Pizza", "Extra Cheese", 1.50, 40, 80.0, 60.00, 50, "Alice").
				AddRow(1, "Pizza", "Peppers", 0.00, 10, 20.0, 0.00, 50, "Bob"))

			w := runDashboardRequest(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var stats DashboardStats
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
			assert.Len(t, stats.ModifierPopularity, 1)
			mods := stats.ModifierPopularity[0].Modifiers
			assert.Equal(t, "Extra Cheese", mods[0].ModifierName)
			assert.Equal(t, "Peppers", mods[1].ModifierName)
			assert.True(t, mods[0].TimesOrdered >= mods[1].TimesOrdered)

			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Produces a warning when modifier popularity query fails", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			mockRevenueSummary(m, 100.00, 100.00, 100.00, 2, 2, 1)
			mockStatusCounts(m, sqlmock.NewRows(statusCols).AddRow("delivered", 2))
			mockBestSelling(m, sqlmock.NewRows(bestSellingCols))
			mockTopCustomers(m, sqlmock.NewRows(topCustomerCols))
			mockEmptySalesTrend(m)
			m.ExpectQuery(modifierPopularityQueryRegex).WillReturnError(fmt.Errorf("db error"))

			w := runDashboardRequest(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var stats DashboardStats
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
			assert.Empty(t, stats.ModifierPopularity)
			assert.Contains(t, stats.Warnings, "modifier popularity unavailable")

			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}
