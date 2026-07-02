// saga_repository.go
// Слой доступа к данным для сохранения и чтения состояния Саги из PostgreSQL.

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SagaRepository определяет интерфейс для работы с хранилищем логов Саги.
type SagaRepository interface {
	SaveSagaLog(ctx context.Context, orderID, stepName, status string) error
	UpdateSagaLogStatus(ctx context.Context, orderID, stepName, newStatus string) error
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	GetSuccessfulSteps(ctx context.Context, orderID string) ([]string, error)
}

// PostgresSagaRepository — реализация репозитория для PostgreSQL.
type PostgresSagaRepository struct {
	db *pgxpool.Pool
}

// NewPostgresSagaRepository создает новый экземпляр репозитория.
func NewPostgresSagaRepository(db *pgxpool.Pool) *PostgresSagaRepository {
	return &PostgresSagaRepository{db: db}
}

// SaveSagaLog сохраняет новую запись в журнал Саги.
func (r *PostgresSagaRepository) SaveSagaLog(ctx context.Context, orderID, stepName, status string) error {
	query := `INSERT INTO saga_logs (order_id, step_name, step_status) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, orderID, stepName, status)
	if err != nil {
		return fmt.Errorf("не удалось сохранить лог саги: %w", err)
	}
	return nil
}

// UpdateSagaLogStatus обновляет статус последнего лога для данного шага.
func (r *PostgresSagaRepository) UpdateSagaLogStatus(ctx context.Context, orderID, stepName, newStatus string) error {
	// Обновляем самый последний лог для этого шага и заказа.
	query := `
        UPDATE saga_logs 
        SET step_status = $1 
        WHERE id = (
            SELECT id FROM saga_logs 
            WHERE order_id = $2 AND step_name = $3 
            ORDER BY created_at DESC 
            LIMIT 1
        )`
	_, err := r.db.Exec(ctx, query, newStatus, orderID, stepName)
	
	if err != nil {
		return fmt.Errorf("не удалось обновить статус лога саги: %w", err)
	}
	return nil
}

// UpdateOrderStatus обновляет статус заказа в таблице orders.
func (r *PostgresSagaRepository) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус заказа: %w", err)
	}
	return nil
}

// GetSuccessfulSteps возвращает имена всех успешно завершенных шагов для данного заказа в обратном порядке их выполнения.
func (r *PostgresSagaRepository) GetSuccessfulSteps(ctx context.Context, orderID string) ([]string, error) {
	query := `
        SELECT step_name 
        FROM saga_logs 
        WHERE order_id = $1 AND step_status = 'SUCCESS' 
        ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить успешные шаги: %w", err)
	}
	defer rows.Close()

	var steps []string
	for rows.Next() {
		var stepName string
		if err := rows.Scan(&stepName); err != nil {
			return nil, fmt.Errorf("ошибка сканирования шага: %w", err)
		}
		steps = append(steps, stepName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return steps, nil
}
