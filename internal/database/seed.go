package database

import (
	"log"
)

func Seed() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)
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

		`INSERT INTO items (name, description, price, category) VALUES ('Margherita Pizza', 'Classic tomato and mozzarella', 12.99, 'Pizza')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Pepperoni Pizza', 'Tomato, mozzarella, pepperoni', 14.99, 'Pizza')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Caesar Salad', 'Romaine, parmesan, croutons', 8.99, 'Salad')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Classic Burger', 'Beef patty, lettuce, tomato', 10.99, 'Burgers')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Chicken Sandwich', 'Grilled chicken, mayo', 9.99, 'Sandwiches')`,
		`INSERT INTO items (name, description, price, category) VALUES ('French Fries', 'Crispy golden fries', 4.99, 'Sides')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Salmon Roll', 'Fresh salmon, rice, nori', 15.99, 'Rolls')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Tuna Roll', 'Fresh tuna, avocado', 14.99, 'Rolls')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Miso Soup', 'Traditional Japanese soup', 3.99, 'Soup')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Cola', 'Classic cola drink', 2.49, 'Drinks')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Orange Juice', 'Fresh orange juice', 3.49, 'Drinks')`,
		`INSERT INTO items (name, description, price, category) VALUES ('Water', 'Bottled water', 1.99, 'Drinks')`,

		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes, scheduled_date) VALUES (1, '123 Main St', 'pending', 25.99, 'Ring doorbell', CURRENT_TIMESTAMP AT TIME ZONE 'UTC')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes, scheduled_date) VALUES (2, '456 Oak Ave', 'preparing', 18.50, '', (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') + INTERVAL '1 day')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes) VALUES (3, '789 Pine Rd', 'delivered', 32.00, 'Leave at door')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes, scheduled_date) VALUES (4, '321 Elm St', 'ready', 15.99, '', (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') + INTERVAL '3 days')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes) VALUES (5, '654 Maple Ave', 'cancelled', 0, 'Customer requested cancellation')`,
		`INSERT INTO orders (customer_id, delivery_address, status, total_amount, notes, scheduled_date) VALUES (1, '123 Main St', 'pending', 12.99, '', (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '1 day')`,

		`INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal) VALUES (1, 1, 2, 12.99, 25.98)`,
		`INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal) VALUES (2, 4, 1, 10.99, 10.99)`,
		`INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal) VALUES (2, 6, 1, 4.99, 4.99)`,
		`INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal) VALUES (2, 10, 1, 2.49, 2.49)`,
		`INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal) VALUES (3, 7, 3, 15.99, 47.97)`,
		`INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal) VALUES (4, 2, 1, 14.99, 14.99)`,
		`INSERT INTO order_items (order_id, item_id, quantity, unit_price, subtotal) VALUES (6, 1, 1, 12.99, 12.99)`,
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
