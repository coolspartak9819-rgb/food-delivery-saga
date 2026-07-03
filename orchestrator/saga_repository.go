package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SagaRepository interface {
	SaveSagaLog(ctx context.Context, orderID, stepName, status string) error
	UpdateSagaLogStatus(ctx context.Context, orderID, stepName, newStatus string) error
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	GetSuccessfulSteps(ctx context.Context, orderID string) ([]string, error)
}

type PostgresSagaRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSagaRepository(db *pgxpool.Pool) *PostgresSagaRepository {
	return &PostgresSagaRepository{db: db}
}

func (r *PostgresSagaRepository) SaveSagaLog(ctx context.Context, orderID, stepName, status string) error {
	query := `INSERT INTO saga_logs (order_id, step_name, step_status) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, orderID, stepName, status)
	return err
}

func (r *PostgresSagaRepository) UpdateSagaLogStatus(ctx context.Context, orderID, stepName, newStatus string) error {
	query := `UPDATE saga_logs SET step_status = $1 WHERE id = (
            SELECT id FROM saga_logs 
            WHERE order_id = $2 AND step_name = $3 
            ORDER BY created_at DESC LIMIT 1)`
	_, err := r.db.Exec(ctx, query, newStatus, orderID, stepName)
	return err
}

func (r *PostgresSagaRepository) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, orderID)
	return err
}

func (r *PostgresSagaRepository) GetSuccessfulSteps(ctx context.Context, orderID string) ([]string, error) {
	query := `SELECT step_name FROM saga_logs WHERE order_id = $1 AND step_status = 'SUCCESS' ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []string
	for rows.Next() {
		var stepName string
		if err := rows.Scan(&stepName); err != nil {
			return nil, err
		}
		steps = append(steps, stepName)
	}
	return steps, rows.Err()
}
