package domain

import (
	"errors"
	"time"
)

type Order struct {
	ID              int64       `json:"id"`
	Number          string      `json:"order_number"`
	CustomerName    string      `json:"customer_name"`
	Type            string      `json:"order_type"`
	TableNumber     int         `json:"table_number"`
	DeliveryAddress string      `json:"delivery_address"`
	TotalAmount     float64     `json:"total_amount"`
	Priority        int         `json:"priority"`
	Status          string      `json:"status"`
	ProcessedBy     string      `json:"processed_by"`
	CompletedAt     time.Time   `json:"completed_at"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items"`
}

type OrderItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	OrderID   int       `json:"order_id"`
}

type Result struct {
	OrderNumber string  `json:"order_number"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"total_amount"`
}

var (
	ErrorBadRequest     = errors.New("Incorrect input")
	InternalServerError = errors.New("Internal Server Error")
)
