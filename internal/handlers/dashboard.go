package handlers

import (
	"log"
	"net/http"
	"time"

	"food-order-tracking/internal/database"

	"github.com/gin-gonic/gin"
)

const (
	bestSellingItemsLimit = 10
	topCustomersLimit     = 5
	salesTrendDays        = 30
)

// DashboardStats is the top-level response for the dashboard endpoint.
type DashboardStats struct {
	TotalRevenue      float64           `json:"total_revenue"`
	MonthlyRevenue    float64           `json:"monthly_revenue"`
	DailyRevenue      float64           `json:"daily_revenue"`
	TotalOrders       int               `json:"total_orders"`
	MonthlyOrders     int               `json:"monthly_orders"`
	DailyOrders       int               `json:"daily_orders"`
	AverageOrderValue float64           `json:"average_order_value"`
	OrdersByStatus    map[string]int    `json:"orders_by_status"`
	BestSellingItems  []BestSellingItem `json:"best_selling_items"`
	TopCustomers      []TopCustomer     `json:"top_customers"`
	SalesTrend        []SalesDataPoint  `json:"sales_trend"`
	Warnings          []string          `json:"warnings,omitempty"`
}

// BestSellingItem holds aggregated sales data for a single menu item.
type BestSellingItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

// TopCustomer holds aggregated order data for a single customer.
type TopCustomer struct {
	Name       string  `json:"name"`
	OrderCount int     `json:"order_count"`
	TotalSpent float64 `json:"total_spent"`
}

// SalesDataPoint holds order count and revenue for a single calendar day.
type SalesDataPoint struct {
	Date    string  `json:"date"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
}

// fetchRevenueSummary populates total, monthly, and daily revenue/order counts
// in a single query each, scoped to the provided time boundaries.
func fetchRevenueSummary(stats *DashboardStats, startOfMonth, startOfDay time.Time) error {
	if err := database.DB.QueryRow(
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM orders WHERE status != 'cancelled'`,
	).Scan(&stats.TotalRevenue, &stats.TotalOrders); err != nil {
		return err
	}

	if err := database.DB.QueryRow(
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM orders WHERE status != 'cancelled' AND created_at >= $1`,
		startOfMonth,
	).Scan(&stats.MonthlyRevenue, &stats.MonthlyOrders); err != nil {
		return err
	}

	if err := database.DB.QueryRow(
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM orders WHERE status != 'cancelled' AND created_at >= $1`,
		startOfDay,
	).Scan(&stats.DailyRevenue, &stats.DailyOrders); err != nil {
		return err
	}

	if stats.TotalOrders > 0 {
		stats.AverageOrderValue = stats.TotalRevenue / float64(stats.TotalOrders)
	}

	return nil
}

// fetchOrdersByStatus populates the OrdersByStatus map.
func fetchOrdersByStatus(stats *DashboardStats) error {
	rows, err := database.DB.Query(
		`SELECT status, COUNT(*) FROM orders
		 WHERE status IS NOT NULL AND status != ''
		 GROUP BY status`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			log.Printf("Error scanning order status row: %v", err)
			continue
		}
		stats.OrdersByStatus[status] = count
	}
	return rows.Err()
}

// fetchBestSellingItems populates the BestSellingItems slice.
func fetchBestSellingItems(stats *DashboardStats) error {
	rows, err := database.DB.Query(`
		SELECT i.name, SUM(oi.quantity) AS total_qty, SUM(oi.subtotal) AS total_revenue
		FROM order_items oi
		JOIN items i ON oi.item_id = i.id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status != 'cancelled'
		GROUP BY i.id, i.name
		ORDER BY total_qty DESC
		LIMIT $1
	`, bestSellingItemsLimit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item BestSellingItem
		if err := rows.Scan(&item.Name, &item.Quantity, &item.Revenue); err != nil {
			log.Printf("Error scanning best selling item row: %v", err)
			continue
		}
		stats.BestSellingItems = append(stats.BestSellingItems, item)
	}
	return rows.Err()
}

// fetchTopCustomers populates the TopCustomers slice.
func fetchTopCustomers(stats *DashboardStats) error {
	rows, err := database.DB.Query(`
		SELECT COALESCE(c.name, 'Unknown'), COUNT(*), COALESCE(SUM(o.total_amount), 0)
		FROM orders o
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE o.status != 'cancelled'
		GROUP BY o.customer_id, c.name
		ORDER BY COUNT(*) DESC
		LIMIT $1
	`, topCustomersLimit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var customer TopCustomer
		if err := rows.Scan(&customer.Name, &customer.OrderCount, &customer.TotalSpent); err != nil {
			log.Printf("Error scanning top customer row: %v", err)
			continue
		}
		stats.TopCustomers = append(stats.TopCustomers, customer)
	}
	return rows.Err()
}

// fetchSalesTrend populates the SalesTrend slice for the past salesTrendDays days
// using a single GROUP BY query instead of one query per day.
func fetchSalesTrend(stats *DashboardStats, since, now time.Time) error {
	rows, err := database.DB.Query(`
		SELECT DATE(created_at AT TIME ZONE 'UTC')::text AS day,
		       COUNT(*) AS orders,
		       COALESCE(SUM(total_amount), 0) AS revenue
		FROM orders
		WHERE status != 'cancelled' AND created_at >= $1
		GROUP BY day
		ORDER BY day ASC
	`, since)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Index query results by date string for O(1) lookup when filling the grid.
	type dayData struct {
		orders  int
		revenue float64
	}
	byDate := make(map[string]dayData, salesTrendDays)
	for rows.Next() {
		var date string
		var d dayData
		if err := rows.Scan(&date, &d.orders, &d.revenue); err != nil {
			log.Printf("Error scanning sales trend row: %v", err)
			continue
		}
		byDate[date] = d
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Build a contiguous day-by-day slice, filling zeros for days with no orders.
	for i := salesTrendDays - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		d := byDate[date] // zero value if date absent
		stats.SalesTrend = append(stats.SalesTrend, SalesDataPoint{
			Date:    date,
			Orders:  d.orders,
			Revenue: d.revenue,
		})
	}
	return nil
}

// GetDashboardStats returns aggregated statistics for the dashboard page.
func GetDashboardStats(c *gin.Context) {
	now := time.Now().UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	salesTrendSince := now.AddDate(0, 0, -(salesTrendDays - 1))

	stats := DashboardStats{
		OrdersByStatus:   make(map[string]int),
		BestSellingItems: make([]BestSellingItem, 0),
		TopCustomers:     make([]TopCustomer, 0),
		SalesTrend:       make([]SalesDataPoint, 0),
	}

	if err := fetchRevenueSummary(&stats, startOfMonth, startOfDay); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := fetchOrdersByStatus(&stats); err != nil {
		log.Printf("Error fetching orders by status: %v", err)
		stats.Warnings = append(stats.Warnings, "orders by status unavailable")
	}
	if err := fetchBestSellingItems(&stats); err != nil {
		log.Printf("Error fetching best selling items: %v", err)
		stats.Warnings = append(stats.Warnings, "best selling items unavailable")
	}
	if err := fetchTopCustomers(&stats); err != nil {
		log.Printf("Error fetching top customers: %v", err)
		stats.Warnings = append(stats.Warnings, "top customers unavailable")
	}
	if err := fetchSalesTrend(&stats, salesTrendSince, now); err != nil {
		log.Printf("Error fetching sales trend: %v", err)
		stats.Warnings = append(stats.Warnings, "sales trend unavailable")
	}

	c.JSON(http.StatusOK, stats)
}
