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
		payment_method VARCHAR(50) DEFAULT 'cash',
		scheduled_date TIMESTAMP,
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
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC');
			RETURN NEW;
		END;
		$$ language 'plpgsql'`,
		`DROP TRIGGER IF EXISTS update_orders_updated_at ON orders`,
		`CREATE TRIGGER update_orders_updated_at
		BEFORE UPDATE ON orders
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column()`,
		`DROP TRIGGER IF EXISTS update_customers_updated_at ON customers`,
		`CREATE TRIGGER update_customers_updated_at
		BEFORE UPDATE ON customers
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column()`,
		`DROP TRIGGER IF EXISTS update_items_updated_at ON items`,
		`CREATE TRIGGER update_items_updated_at
		BEFORE UPDATE ON items
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column()`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS scheduled_date TIMESTAMP`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT 'cash'`,
		`ALTER TABLE orders ALTER COLUMN scheduled_date TYPE TIMESTAMP`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT 'cash'`,
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
