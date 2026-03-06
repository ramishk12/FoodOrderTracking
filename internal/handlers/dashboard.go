package handlers

import (
	"net/http"
	"time"

	"food-order-tracking/internal/database"

	"github.com/gin-gonic/gin"
)

type DashboardStats struct {
	TotalRevenue     float64           `json:"total_revenue"`
	MonthlyRevenue   float64           `json:"monthly_revenue"`
	DailyRevenue     float64           `json:"daily_revenue"`
	TotalOrders     int                `json:"total_orders"`
	MonthlyOrders   int                `json:"monthly_orders"`
	DailyOrders     int                `json:"daily_orders"`
	AverageOrderValue float64          `json:"average_order_value"`
	OrdersByStatus  map[string]int     `json:"orders_by_status"`
	BestSellingItems []BestSellingItem `json:"best_selling_items"`
	TopCustomers    []TopCustomer       `json:"top_customers"`
	SalesTrend      []SalesDataPoint    `json:"sales_trend"`
}

type BestSellingItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

type TopCustomer struct {
	Name      string  `json:"name"`
	OrderCount int    `json:"order_count"`
	TotalSpent float64 `json:"total_spent"`
}

type SalesDataPoint struct {
	Date   string  `json:"date"`
	Orders int     `json:"orders"`
	Revenue float64 `json:"revenue"`
}

func GetDashboardStats(c *gin.Context) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var stats DashboardStats
	stats.OrdersByStatus = make(map[string]int)

	// Total revenue and orders (all time)
	err := database.DB.QueryRow("SELECT COALESCE(SUM(total_amount), 0), COUNT(*) FROM orders WHERE status != 'cancelled'").Scan(&stats.TotalRevenue, &stats.TotalOrders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Monthly revenue and orders
	err = database.DB.QueryRow("SELECT COALESCE(SUM(total_amount), 0), COUNT(*) FROM orders WHERE status != 'cancelled' AND created_at >= $1", startOfMonth).Scan(&stats.MonthlyRevenue, &stats.MonthlyOrders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Daily revenue and orders
	err = database.DB.QueryRow("SELECT COALESCE(SUM(total_amount), 0), COUNT(*) FROM orders WHERE status != 'cancelled' AND created_at >= $1", startOfDay).Scan(&stats.DailyRevenue, &stats.DailyOrders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Average order value
	if stats.TotalOrders > 0 {
		stats.AverageOrderValue = stats.TotalRevenue / float64(stats.TotalOrders)
	}

	// Orders by status
	statusRows, err := database.DB.Query("SELECT status, COUNT(*) FROM orders GROUP BY status")
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var count int
			statusRows.Scan(&status, &count)
			stats.OrdersByStatus[status] = count
		}
	}

	// Best selling items
	itemRows, err := database.DB.Query(`
		SELECT i.name, SUM(oi.quantity) as total_qty, SUM(oi.subtotal) as total_revenue
		FROM order_items oi
		JOIN items i ON oi.item_id = i.id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status != 'cancelled'
		GROUP BY i.id, i.name
		ORDER BY total_qty DESC
		LIMIT 10
	`)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var item BestSellingItem
			itemRows.Scan(&item.Name, &item.Quantity, &item.Revenue)
			stats.BestSellingItems = append(stats.BestSellingItems, item)
		}
	}

	// Top customers
	customerRows, err := database.DB.Query(`
		SELECT COALESCE(c.name, 'Unknown'), COUNT(*), SUM(o.total_amount)
		FROM orders o
		LEFT JOIN customers c ON o.customer_id = c.id
		WHERE o.status != 'cancelled'
		GROUP BY o.customer_id, c.name
		ORDER BY COUNT(*) DESC
		LIMIT 5
	`)
	if err == nil {
		defer customerRows.Close()
		for customerRows.Next() {
			var customer TopCustomer
			customerRows.Scan(&customer.Name, &customer.OrderCount, &customer.TotalSpent)
			stats.TopCustomers = append(stats.TopCustomers, customer)
		}
	}

	// Sales trend - last 30 days
	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDate := startOfDate.AddDate(0, 0, 1)

		var revenue float64
		var orders int
		database.DB.QueryRow(`
			SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
			FROM orders
			WHERE status != 'cancelled' AND created_at >= $1 AND created_at < $2
		`, startOfDate, endOfDate).Scan(&revenue, &orders)

		stats.SalesTrend = append(stats.SalesTrend, SalesDataPoint{
			Date:    startOfDate.Format("2006-01-02"),
			Orders:  orders,
			Revenue: revenue,
		})
	}

	c.JSON(http.StatusOK, stats)
}
