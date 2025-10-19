package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"wheres-my-pizza/order-service/internal/core/domain"
	"wheres-my-pizza/order-service/internal/infrastructure/config"
)

type postgres struct {
	db *sql.DB
}

func NewPostgresRepository(config *config.DB) *postgres {
	url := fmt.Sprintf("user=%s password=%s host=%s port=%s database=%s sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.Name,
	)

	db, err := sql.Open("pgx", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v", err)
	}

	return &postgres{
		db: db,
	}
}

func (postgresRepo *postgres) CreateOrder(order domain.Order) error {
	tx, err := postgresRepo.db.Begin()
	if err != nil {
		return fmt.Errorf("Error Transaction Begin: %v", err)
	}
	defer tx.Rollback()

	var query = `INSERT INTO orders (number, customer_name, type, total_amount) VALUES($1, $2, $3, $4) RETURNING id`
	var orderID int64

	err = tx.QueryRow(query, order.Number, order.CustomerName, order.Type, order.TotalAmount).Scan(&orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("Error: create order: %v", err)
		}
		return fmt.Errorf("Error: create order: %v", err)
	}

	for i := range order.Items {
		query = `INSERT INTO order_items (name, quantity, price, order_id) VALUES($1, $2, $3, $4)`

		_, err = tx.Exec(query, order.Items[i].Name, order.Items[i].Quantity, order.Items[i].Price, orderID)
		if err != nil {
			return fmt.Errorf("Error: write order items: %v", err)
		}
	}

	query = `INSERT INTO order_status_log (status, order_id) VALUES($1, $2)`

	_, err = tx.Exec(query, order.Status, orderID)
	if err != nil {
		return fmt.Errorf("Error: write order status: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("Error Transaction Commit: %v", err)
	}

	return nil
}

func (postgresRepo *postgres) GetOrderNumber() (string, error) {
	var order domain.Order

	row := postgresRepo.db.QueryRow("SELECT number FROM orders ORDER BY id DESC LIMIT 1")
	if err := row.Scan(&order.Number); err != nil {
		if err == sql.ErrNoRows {
			return order.Number, fmt.Errorf("Error: %v", err)
		}
		return order.Number, fmt.Errorf("Error: %v", err)
	}

	return order.Number, nil
}
