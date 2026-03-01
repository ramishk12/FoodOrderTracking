package models

import "time"

type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
	ID              int       `json:"id"`
	CustomerID      int       `json:"customer_id"`
	CustomerName    string    `json:"customer_name"`
	CustomerPhone   string    `json:"customer_phone"`
	DeliveryAddress string    `json:"delivery_address"`
	Status          string    `json:"status"`
	TotalAmount     float64   `json:"total_amount"`
	Items           string    `json:"items"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
