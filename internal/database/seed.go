package database

import (
	"log"
)

func Seed() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM customers").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("Database already seeded")
		return nil
	}

	seed := []string{
		`INSERT INTO customers (name, phone, email, address) VALUES ('John Doe', '555-1234', 'john@example.com', '123 Main St')`,
		`INSERT INTO customers (name, phone, email, address) VALUES ('Jane Smith', '555-5678', 'jane@example.com', '456 Oak Ave')`,
		`INSERT INTO customers (name, phone, email, address) VALUES ('Bob Wilson', '555-9012', 'bob@example.com', '789 Pine Rd')`,
		`INSERT INTO customers (name, phone, email, address) VALUES ('Alice Brown', '555-3456', 'alice@example.com', '321 Elm St')`,
		`INSERT INTO customers (name, phone, email, address) VALUES ('Charlie Davis', '555-7890', 'charlie@example.com', '654 Maple Ave')`,

		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, items, notes) VALUES (1, '123 Main St', 'pending', 25.99, '2x Margherita Pizza, 1x Caesar Salad', 'Ring doorbell')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, items, notes) VALUES (2, '456 Oak Ave', 'preparing', 18.50, '1x Classic Burger, 1x French Fries', '')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, items, notes) VALUES (3, '789 Pine Rd', 'delivered', 32.00, '3x Salmon Roll, 1x Miso Soup', 'Leave at door')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, items, notes) VALUES (4, '321 Elm St', 'ready', 15.99, '1x Pepperoni Pizza', '')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, items, notes) VALUES (5, '654 Maple Ave', 'cancelled', 0, '', 'Customer requested cancellation')`,
	}

	for _, s := range seed {
		if _, err := DB.Exec(s); err != nil {
			log.Printf("Seed error: %v", err)
			return err
		}
	}

	log.Println("Database seeded successfully")
	return nil
}
