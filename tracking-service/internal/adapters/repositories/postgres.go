package repositories

import (
	"database/sql"
	"fmt"
	"wheres-my-pizza/tracking-service/internal/core/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgres struct {
	Db *sql.DB
}

func NewUserRepository(cfg *domain.DatabaseConfig) (*postgres, error) {
	url := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &postgres{Db: db}, nil
}

func (r *postgres) GetOrder(orderNum string) (*domain.Order, error) {
	var order domain.Order
	row := r.Db.QueryRow("SELECT number, status, updated_at, completed_at, processed_by from orders where number = $1", orderNum)

	if err := row.Scan(&order.Order_number, &order.Current_status, &order.Updated_at, &order.Estimated_completion, &order.Processed_by); err != nil {
		return &order, err
	}

	return &order, nil
}

func (r *postgres) GetOrderStatus(orderNum string) (*[]domain.Status, error) {
	var orderStatus []domain.Status
	rows, err := r.Db.Query("SELECT s.status, s.changed_at, s.changed_by from order_status_log as s inner join orders on orders.id = s.order_id where orders.id in (select id from orders where number = $1)", orderNum)

	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var order domain.Status
		if err := rows.Scan(&order.Status, &order.Timestamp, &order.Changed_by); err != nil {
			return nil, err
		}
		orderStatus = append(orderStatus, order)
	}

	return &orderStatus, nil

}

func (r *postgres) GetWorkersStatus() (*[]domain.Worker, error) {
	var workers []domain.Worker
	rows, err := r.Db.Query("SELECT name,orders_processed,last_seen, case when now() - last_seen > interval '60 seconds' then 'offline' else 'online' end as status from workers")

	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var worker domain.Worker
		if err := rows.Scan(&worker.Worker_name, &worker.Orders_processed, &worker.Last_seen, &worker.Status); err != nil {
			return nil, err
		}
		workers = append(workers, worker)
	}

	return &workers, nil
}
