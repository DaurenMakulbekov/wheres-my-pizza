package postgres

import (
	"database/sql"
	"fmt"
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
	var query = `INSERT INTO workers (name, type, created_at, last_seen) VALUES($1, $2, $3, $4)`

	_, err := database.db.Exec(query, worker.Name, worker.Type, worker.CreatedAt, worker.LastSeen)
	if err != nil {
		return fmt.Errorf("Error: write worker: %v", err)
	}

	return nil
}
