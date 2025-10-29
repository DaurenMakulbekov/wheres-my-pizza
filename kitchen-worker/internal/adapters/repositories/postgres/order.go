package postgres

import (
	"database/sql"
	"fmt"
	"time"
	"wheres-my-pizza/kitchen-worker/internal/core/domain"
	"wheres-my-pizza/kitchen-worker/internal/infrastructure/config"
)

type database struct {
	db *sql.DB
}

func ConnectDB(config *config.DB) (*sql.DB, error) {
	url := fmt.Sprintf("user=%s password=%s host=%s port=%s database=%s sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.Name,
	)

	db, err := sql.Open("pgx", url)
	if err != nil {
		return db, err
	}

	if err = db.Ping(); err != nil {
		return db, err
	}

	return db, nil
}

func NewDatabaseRepository(db *sql.DB) *database {
	return &database{
		db: db,
	}
}

func (database *database) Register(worker domain.Worker) error {
	var query = `INSERT INTO workers (name, type) VALUES($1, $2)`

	_, err := database.db.Exec(query, worker.Name, worker.Type)
	if err != nil {
		return fmt.Errorf("Error: write worker: %v", err)
	}

	return nil
}

func (database *database) GetWorker(name string) (domain.Worker, error) {
	var worker domain.Worker
	var query = `SELECT name, type, status, orders_processed, created_at, last_seen FROM workers WHERE name = $1`

	var row = database.db.QueryRow(query, name)
	if err := row.Scan(&worker.Name, &worker.Type, &worker.Status, &worker.OrdersProcessed, &worker.CreatedAt, &worker.LastSeen); err != nil {
		if err == sql.ErrNoRows {
			return worker, fmt.Errorf("Error: get worker: %v", err)
		}
		return worker, fmt.Errorf("Error: get worker: %v", err)
	}

	return worker, nil
}

func (database *database) UpdateWorker(worker domain.Worker) error {
	var query = `UPDATE workers SET status = $1, type = $2 WHERE name = $3`

	_, err := database.db.Exec(query, worker.Status, worker.Type, worker.Name)
	if err != nil {
		return fmt.Errorf("Error: update worker status: %v", err)
	}

	return nil
}

func (database *database) UpdateOrder(worker domain.Worker, order domain.Order) error {
	tx, err := database.db.Begin()
	if err != nil {
		return fmt.Errorf("Error Transaction Begin: %v", err)
	}
	defer tx.Rollback()

	var query = `UPDATE orders SET status = 'cooking', processed_by = $1, updated_at = $2 WHERE number = $3`

	_, err = tx.Exec(query, worker.Name, time.Now(), order.Number)
	if err != nil {
		return fmt.Errorf("Error: update order: %v", err)
	}

	query = `UPDATE order_status_log SET status = 'cooking', changed_by = $1, changed_at = $2 WHERE order_id = $3`

	_, err = tx.Exec(query, worker.Name, time.Now(), order.ID)
	if err != nil {
		return fmt.Errorf("Error: update order status: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("Error Transaction Commit: %v", err)
	}

	return nil
}

func (database *database) UpdateOrderReady(worker domain.Worker, order domain.Order) error {
	tx, err := database.db.Begin()
	if err != nil {
		return fmt.Errorf("Error Transaction Begin: %v", err)
	}
	defer tx.Rollback()

	var query = `UPDATE orders SET status = 'ready', completed_at = $1 WHERE number = $2`

	_, err = tx.Exec(query, time.Now(), order.Number)
	if err != nil {
		return fmt.Errorf("Error: update order: %v", err)
	}

	query = `UPDATE order_status_log SET status = 'ready', changed_at = $1 WHERE order_id = $2`

	_, err = tx.Exec(query, time.Now(), order.ID)
	if err != nil {
		return fmt.Errorf("Error: update order status: %v", err)
	}

	query = `UPDATE workers SET orders_processed += 1 WHERE name = $1`

	_, err = tx.Exec(query, worker.Name)
	if err != nil {
		return fmt.Errorf("Error: update worker 'orders_processed': %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("Error Transaction Commit: %v", err)
	}

	return nil
}
