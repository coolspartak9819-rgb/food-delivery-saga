package main

import (
	"context"
	"fmt"
	"log"
)

type SagaStep interface {
	Name() string
	Execute(ctx context.Context, orderID string) error
	Compensate(ctx context.Context, orderID string) error
}

type SagaOrchestrator struct {
	steps      []SagaStep
	repository SagaRepository
}

func NewSagaOrchestrator(steps []SagaStep, repo SagaRepository) *SagaOrchestrator {
	return &SagaOrchestrator{steps: steps, repository: repo}
}

func (o *SagaOrchestrator) ExecuteSaga(ctx context.Context, orderID string) error {
	stepMap := make(map[string]SagaStep)
	for _, s := range o.steps {
		stepMap[s.Name()] = s
	}

	for _, step := range o.steps {
		log.Printf("ОРКЕСТРАТОР: Выполняется шаг '%s' для заказа %s", step.Name(), orderID)
		if err := o.repository.SaveSagaLog(ctx, orderID, step.Name(), "STARTED"); err != nil {
			return fmt.Errorf("DB log error: %w", err)
		}

		err := step.Execute(ctx, orderID)
		if err != nil {
			log.Printf("ОРКЕСТРАТОР: Ошибка на шаге '%s': %v. Запуск компенсации...", step.Name(), err)
			_ = o.repository.UpdateSagaLogStatus(ctx, orderID, step.Name(), "FAILED")
			o.compensate(ctx, orderID, stepMap)
			return fmt.Errorf("сага прервана на шаге '%s': %w", step.Name(), err)
		}

		if err := o.repository.UpdateSagaLogStatus(ctx, orderID, step.Name(), "SUCCESS"); err != nil {
			return fmt.Errorf("DB log error: %w", err)
		}
	}

	log.Printf("ОРКЕСТРАТОР: Сага для заказа %s успешно завершена.", orderID)
	return o.repository.UpdateOrderStatus(ctx, orderID, "COMPLETED")
}

func (o *SagaOrchestrator) compensate(ctx context.Context, orderID string, stepMap map[string]SagaStep) {
	successfulSteps, err := o.repository.GetSuccessfulSteps(ctx, orderID)
	if err != nil {
		log.Printf("КРИТИЧЕСКАЯ ОШИБКА: не удалось получить шаги для компенсации: %v", err)
		return
	}

	for _, stepName := range successfulSteps {
		if compStep, ok := stepMap[stepName]; ok {
			log.Printf("ОРКЕСТРАТОР: Компенсация шага '%s' для заказа %s", compStep.Name(), orderID)
			if err := compStep.Compensate(ctx, orderID); err != nil {
				log.Printf("ВНИМАНИЕ: Не удалось компенсировать шаг '%s': %v", compStep.Name(), err)
				_ = o.repository.UpdateSagaLogStatus(ctx, orderID, compStep.Name(), "COMPENSATION_FAILED")
			} else {
				_ = o.repository.UpdateSagaLogStatus(ctx, orderID, compStep.Name(), "COMPENSATED")
			}
		}
	}
	_ = o.repository.UpdateOrderStatus(ctx, orderID, "CANCELLED")
}
