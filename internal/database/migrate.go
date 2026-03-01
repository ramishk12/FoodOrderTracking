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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
			delivery_address TEXT NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			total_amount DECIMAL(10,2) NOT NULL,
			items TEXT,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
