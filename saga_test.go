package main

import (
	"context"
	"strings"
	"testing"
)

// TestSaga_SuccessFlow проверяет успешный сценарий, когда все три сервиса отрабатывают без ошибок.
func TestSaga_SuccessFlow(t *testing.T) {
	t.Log("=== Запуск теста: Успешное выполнение саги ===")

	orderSvc := &OrderService{}
	paymentSvc := &PaymentService{}
	kitchenSvc := &KitchenService{}

	steps := []SagaStep{orderSvc, paymentSvc, kitchenSvc}
	orchestrator := NewSagaOrchestrator(steps)

	ctx := context.Background()
	orderID := "ORDER-100"

	err := orchestrator.ExecuteSaga(ctx, orderID)

	if err != nil {
		t.Fatalf("Ожидалось успешное выполнение саги, но получена ошибка: %v", err)
	}

	// Проверяем, что Execute был вызван у всех
	if !orderSvc.ExecuteCalled || !paymentSvc.ExecuteCalled || !kitchenSvc.ExecuteCalled {
		t.Errorf("Ожидалось, что все методы Execute будут вызваны")
	}

	// Проверяем, что Compensate НЕ был вызван ни у кого
	if orderSvc.CompensateCalled || paymentSvc.CompensateCalled || kitchenSvc.CompensateCalled {
		t.Errorf("Ожидалось, что методы Compensate не будут вызваны при успешном сценарии")
	}

	t.Log("Тест TestSaga_SuccessFlow пройден успешно.")
}

// TestSaga_CompensatingFlow проверяет сценарий сбоя на этапе кухни и последующий откат (LIFO).
func TestSaga_CompensatingFlow(t *testing.T) {
	t.Log("=== Запуск теста: Сбой саги и вызов компенсаций ===")

	orderSvc := &OrderService{}
	paymentSvc := &PaymentService{}
	// Искусственно вызываем ошибку на шаге кухни
	kitchenSvc := &KitchenService{ShouldFail: true}

	steps := []SagaStep{orderSvc, paymentSvc, kitchenSvc}
	orchestrator := NewSagaOrchestrator(steps)

	ctx := context.Background()
	orderID := "ORDER-404"

	err := orchestrator.ExecuteSaga(ctx, orderID)

	if err == nil {
		t.Fatalf("Ожидалась ошибка от оркестратора, но сага завершилась 'успешно'")
	}

	// Проверяем, что ошибка пришла именно от KitchenService
	expectedSubstr := "закончилось тесто для пиццы"
	if !strings.Contains(err.Error(), expectedSubstr) {
		t.Errorf("Ожидалась ошибка содержащая '%s', получено: %v", expectedSubstr, err)
	}
	t.Logf("Оркестратор успешно перехватил ошибку: %v", err)

	// Проверяем Execute (должны быть вызваны все 3, так как ошибка возникает ВНУТРИ Execute KitchenService)
	if !orderSvc.ExecuteCalled || !paymentSvc.ExecuteCalled || !kitchenSvc.ExecuteCalled {
		t.Errorf("Ожидалось, что все методы Execute будут вызваны (до момента сбоя)")
	}

	// Проверяем Compensate. Должен быть вызван у Payment и Order. У Kitchen НЕ должен вызываться.
	if !paymentSvc.CompensateCalled {
		t.Errorf("Ожидалось, что PaymentService.Compensate будет вызван для отката")
	} else {
		t.Log("PaymentService: компенсация (возврат средств) успешно вызвана.")
	}

	if !orderSvc.CompensateCalled {
		t.Errorf("Ожидалось, что OrderService.Compensate будет вызван для отката")
	} else {
		t.Log("OrderService: компенсация (отмена заказа) успешно вызвана.")
	}

	if kitchenSvc.CompensateCalled {
		t.Errorf("Не ожидалось, что KitchenService.Compensate будет вызван, так как он был источником ошибки")
	}

	t.Log("Тест TestSaga_CompensatingFlow пройден успешно. Откаты отработали.")
}
