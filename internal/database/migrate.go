package database

import (
	"log"
)

func Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS customers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(50),
			email VARCHAR(255),
			address TEXT,
			created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		)`,
		`CREATE TABLE IF NOT EXISTS items (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			category VARCHAR(100),
			available BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
			delivery_address TEXT NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			total_amount DECIMAL(10,2) NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
			item_id INTEGER REFERENCES items(id) ON DELETE SET NULL,
			quantity INTEGER NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			subtotal DECIMAL(10,2) NOT NULL
		)`,
	}

	for _, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			log.Printf("Migration error: %v", err)
			return err
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}
