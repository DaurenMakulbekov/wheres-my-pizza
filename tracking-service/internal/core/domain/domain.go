package domain

import "time"

type Order struct {
	Order_number         string     `json:"order_number"`
	Current_status       string     `json:"current_status"`
	Updated_at           string     `json:"updated_at"`
	Estimated_completion *time.Time `json:"estimated_completion"`
	Processed_by         string     `json:"processed_by"`
}

type Status struct {
	Status     string     `json:"status"`
	Timestamp  *time.Time `json:"timestamp"`
	Changed_by string     `json:"changed_by"`
}

type Worker struct {
	Worker_name      string     `json:"worker_name"`
	Status           string     `json:"status"`
	Orders_processed string     `json:"orders_processed"`
	Last_seen        *time.Time `json:"last_seen"`
}
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}
