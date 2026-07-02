package main

import (
	"context"
	"errors"
	"fmt"
)

// PaymentService отвечает за процесс оплаты.
type PaymentService struct {
	ShouldFail        bool
	ExecuteCalled     bool
	CompensateCalled  bool
}

func (s *PaymentService) Name() string { return "PaymentService" }

func (s *PaymentService) Execute(ctx context.Context, orderID string) error {
	s.ExecuteCalled = true
	if s.ShouldFail {
		return errors.New("недостаточно средств или ошибка шлюза")
	}
	fmt.Printf("[%s] Списание денег для заказа %s прошло успешно\n", s.Name(), orderID)
	return nil
}

func (s *PaymentService) Compensate(ctx context.Context, orderID string) error {
	s.CompensateCalled = true
	fmt.Printf("[%s] Возврат средств на карту пользователя для заказа %s выполнен (Откат)\n", s.Name(), orderID)
	return nil
}

// KitchenService отвечает за процесс приготовления (кухню).
type KitchenService struct {
	ShouldFail        bool
	ExecuteCalled     bool
	CompensateCalled  bool
}

func (s *KitchenService) Name() string { return "KitchenService" }

func (s *KitchenService) Execute(ctx context.Context, orderID string) error {
	s.ExecuteCalled = true
	if s.ShouldFail {
		return errors.New("закончилось тесто для пиццы")
	}
	fmt.Printf("[%s] Ингредиенты для заказа %s зарезервированы на кухне\n", s.Name(), orderID)
	return nil
}

func (s *KitchenService) Compensate(ctx context.Context, orderID string) error {
	s.CompensateCalled = true
	fmt.Printf("[%s] Снятие резерва ингредиентов для заказа %s (Откат)\n", s.Name(), orderID)
	return nil
}
