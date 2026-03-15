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

type Item struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Category    string         `json:"category"`
	Available   bool           `json:"available"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Modifiers   []ItemModifier `json:"modifiers"`
}

type Order struct {
	ID              int         `json:"id"`
	CustomerID      int         `json:"customer_id"`
	CustomerName    string      `json:"customer_name"`
	CustomerPhone   string      `json:"customer_phone"`
	DeliveryAddress string      `json:"delivery_address"`
	Status          string      `json:"status"`
	TotalAmount     float64     `json:"total_amount"`
	Notes           string      `json:"notes"`
	PaymentMethod   string      `json:"payment_method"`
	OrderItems      []OrderItem `json:"order_items"`
	ScheduledDate   *time.Time  `json:"scheduled_date"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// ItemModifier is a modifier that belongs to a specific menu item.
type ItemModifier struct {
	ID              int     `json:"id"`
	ItemID          int     `json:"item_id"`
	Name            string  `json:"name"`
	PriceAdjustment float64 `json:"price_adjustment"`
}

// OrderItemModifier is a modifier applied to a specific order item line.
// ModifierID is 0 when the modifier was entered ad-hoc (not from item_modifiers table).
type OrderItemModifier struct {
	ID              int     `json:"id"`
	OrderItemID     int     `json:"order_item_id"`
	ModifierID      int     `json:"modifier_id,omitempty"`
	ModifierName    string  `json:"modifier_name"`
	PriceAdjustment float64 `json:"price_adjustment"`
}

type OrderItem struct {
	ID        int                 `json:"id"`
	OrderID   int                 `json:"order_id"`
	ItemID    int                 `json:"item_id"`
	ItemName  string              `json:"item_name"`
	Quantity  int                 `json:"quantity"`
	UnitPrice float64             `json:"unit_price"`
	Subtotal  float64             `json:"subtotal"`
	Modifiers []OrderItemModifier `json:"modifiers"`
}
