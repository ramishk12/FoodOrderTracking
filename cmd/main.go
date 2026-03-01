package main

import (
	"log"
	"net/http"

	"food-order-tracking/internal/database"
	"food-order-tracking/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "pgtest",
		DBName:   "food_order_tracking",
	}

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
			orders.GET("/:id", handlers.GetOrder)
			orders.POST("", handlers.CreateOrder)
			orders.PUT("/:id", handlers.UpdateOrder)
			orders.DELETE("/:id", handlers.DeleteOrder)
		}
	}

	log.Println("Server starting on :8080")
	r.Run()
}
