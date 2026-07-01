package main

import (
	"context"
	"fmt"
)

// SagaStep определяет интерфейс для одного шага в саге.
type SagaStep interface {
	Name() string
	Execute(ctx context.Context, orderID string) error
	Compensate(ctx context.Context, orderID string) error
}

// SagaOrchestrator управляет выполнением и компенсацией последовательности шагов саги.
type SagaOrchestrator struct {
	steps []SagaStep
}

// NewSagaOrchestrator создает новый SagaOrchestrator с заданными шагами.
func NewSagaOrchestrator(steps []SagaStep) *SagaOrchestrator {
	return &SagaOrchestrator{
		steps: steps,
	}
}

// ExecuteSaga последовательно выполняет шаги саги.
// Если на каком-то из шагов происходит сбой, оркестратор немедленно прерывает выполнение
// и запускает метод Compensate у всех ранее успешно выполненных шагов в обратном порядке (LIFO).
func (o *SagaOrchestrator) ExecuteSaga(ctx context.Context, orderID string) error {
	var executedSteps []SagaStep

	for _, step := range o.steps {
		err := step.Execute(ctx, orderID)
		if err != nil {
			// Ошибка на шаге: запускаем компенсацию для всех ранее успешных шагов в обратном порядке
			for i := len(executedSteps) - 1; i >= 0; i-- {
				compStep := executedSteps[i]
				_ = compStep.Compensate(ctx, orderID)
				// В реальной системе здесь стоит логировать ошибки компенсации
				// или отправлять их в систему мониторинга (DLQ, Retry механизмы).
			}
			return fmt.Errorf("сбой на шаге %s: %w", step.Name(), err)
		}
		
		// Запоминаем успешно выполненный шаг
		executedSteps = append(executedSteps, step)
	}

	return nil
}
