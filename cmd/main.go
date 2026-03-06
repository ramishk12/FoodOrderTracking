package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Get database config from environment variables with defaults
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "172.17.32.1"
	}

	port := 5432
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "pgtest"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "food_order_tracking"
	}

	cfg := database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}

	log.Printf("Connecting to database at %s:%d as %s", cfg.Host, cfg.Port, cfg.User)

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	database.Seed()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		customers := api.Group("/customers")
		{
			customers.GET("", handlers.GetCustomers)
			customers.GET("/:id", handlers.GetCustomer)
			customers.POST("", handlers.CreateCustomer)
			customers.PUT("/:id", handlers.UpdateCustomer)
			customers.DELETE("/:id", handlers.DeleteCustomer)
		}

		items := api.Group("/items")
		{
			items.GET("", handlers.GetItems)
			items.GET("/:id", handlers.GetItem)
			items.POST("", handlers.CreateItem)
			items.PUT("/:id", handlers.UpdateItem)
			items.DELETE("/:id", handlers.DeleteItem)
		}

		orders := api.Group("/orders")
		{
			orders.GET("", handlers.GetOrders)
			orders.POST("", handlers.CreateOrder)
			orders.GET("/scheduled", handlers.GetScheduledOrders)
			orders.GET("/customer/:customerId", handlers.GetOrdersByCustomer)
			orders.GET("/:id", handlers.GetOrder)
			orders.PUT("/:id", handlers.UpdateOrder)
			orders.DELETE("/:id", handlers.DeleteOrder)
		}

		api.GET("/dashboard", handlers.GetDashboardStats)
	}

	log.Println("Server starting on :8080")
	r.Run()
}
