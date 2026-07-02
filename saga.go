// saga.go
// Обновленное ядро оркестратора с интеграцией персистентного хранилища.

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

// SagaOrchestrator управляет выполнением и компенсацией, используя репозиторий для персистентности.
type SagaOrchestrator struct {
	steps      []SagaStep
	repository SagaRepository
}

// NewSagaOrchestrator создает новый оркестратор с шагами и репозиторием.
func NewSagaOrchestrator(steps []SagaStep, repo SagaRepository) *SagaOrchestrator {
	return &SagaOrchestrator{
		steps:      steps,
		repository: repo,
	}
}

// ExecuteSaga выполняет сагу, логируя каждый шаг в БД.
func (o *SagaOrchestrator) ExecuteSaga(ctx context.Context, orderID string) error {
	// Создаем map для быстрого доступа к шагам по имени во время компенсации
	stepMap := make(map[string]SagaStep)
	for _, s := range o.steps {
		stepMap[s.Name()] = s
	}

	for _, step := range o.steps {
		// 1. Логируем НАМЕРЕНИЕ выполнить шаг
		if err := o.repository.SaveSagaLog(ctx, orderID, step.Name(), "STARTED"); err != nil {
			return fmt.Errorf("критическая ошибка: не удалось залогировать старт шага %s: %w", step.Name(), err)
		}

		// 2. Выполняем сам шаг
		fmt.Printf("[Orchestrator] Выполняется шаг: %s для заказа %s\n", step.Name(), orderID)
		err := step.Execute(ctx, orderID)

		if err != nil {
			// 3a. Ошибка выполнения: логируем сбой и запускаем компенсацию
			fmt.Printf("[Orchestrator] Ошибка на шаге %s: %v. Инициируется процесс компенсации...\n", step.Name(), err)
			_ = o.repository.UpdateSagaLogStatus(ctx, orderID, step.Name(), "FAILED")
			
			o.compensate(ctx, orderID, stepMap)
			
			return fmt.Errorf("сага прервана на шаге %s: %w", step.Name(), err)
		}

		// 3b. Успешное выполнение: обновляем лог
		if err := o.repository.UpdateSagaLogStatus(ctx, orderID, step.Name(), "SUCCESS"); err != nil {
			// Это тоже критическая ситуация, так как мы не сможем откатить этот шаг
			return fmt.Errorf("критическая ошибка: не удалось залогировать успех шага %s: %w", step.Name(), err)
		}
	}

	fmt.Printf("[Orchestrator] Сага успешно завершена для заказа %s\n", orderID)
	return o.repository.UpdateOrderStatus(ctx, orderID, "COMPLETED")
}

// compensate читает из БД успешно выполненные шаги и откатывает их.
func (o *SagaOrchestrator) compensate(ctx context.Context, orderID string, stepMap map[string]SagaStep) {
	successfulSteps, err := o.repository.GetSuccessfulSteps(ctx, orderID)
	if err != nil {
		fmt.Printf("[Orchestrator] КРИТИЧЕСКАЯ ОШИБКА: не удалось получить список шагов для компенсации: %v\n", err)
		// Здесь нужна логика для retry или алертинг в систему мониторинга
		return
	}

	// Список уже отсортирован в LIFO порядке благодаря SQL-запросу
	for _, stepName := range successfulSteps {
		if compStep, ok := stepMap[stepName]; ok {
			fmt.Printf("[Orchestrator] Компенсация шага: %s для заказа %s\n", compStep.Name(), orderID)
			if err := compStep.Compensate(ctx, orderID); err != nil {
				fmt.Printf("[Orchestrator] ВНИМАНИЕ: Не удалось компенсировать шаг %s: %v\n", compStep.Name(), err)
				// Логируем ошибку компенсации, но продолжаем, чтобы откатить остальные
				_ = o.repository.UpdateSagaLogStatus(ctx, orderID, compStep.Name(), "COMPENSATION_FAILED")
			} else {
				// Логируем успешную компенсацию
				_ = o.repository.UpdateSagaLogStatus(ctx, orderID, compStep.Name(), "COMPENSATED")
			}
		}
	}
	// В конце меняем статус всего заказа на отмененный
	_ = o.repository.UpdateOrderStatus(ctx, orderID, "CANCELLED")
}
