package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// Убедитесь, что эта строка соответствует вашим настройкам БД
	dbConnectionString   = "postgres://user:password@localhost:5432/saga_db?sslmode=disable"
	orderServiceAddress = "localhost:50051"
)

func main() {
	ctx := context.Background()

	// 1. Подключение к PostgreSQL
	dbpool, err := pgxpool.New(ctx, dbConnectionString)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbpool.Close()
	log.Println("ОРКЕСТРАТОР: Успешное подключение к PostgreSQL.")

	// 2. Создание репозитория для логирования саги
	sagaRepo := NewPostgresSagaRepository(dbpool)

	// 3. Установка gRPC соединения с OrderService
	// Используем WithBlock, чтобы дождаться установки соединения перед продолжением
	orderConn, err := grpc.Dial(orderServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("ОРКЕСТРАТОР: Не удалось подключиться к OrderService по gRPC: %v", err)
	}
	defer orderConn.Close()
	log.Println("ОРКЕСТРАТОР: Успешное gRPC подключение к OrderService.")

	// 4. Создание шагов Саги
	// Первый шаг - это уже не мок, а настоящий gRPC клиент
	orderStep := NewOrderServiceGRPCClient(orderConn)

	// Остальные шаги пока остаются моками
	paymentStep := &PaymentService{ShouldFail: false} // Успешный сценарий
	kitchenStep := &KitchenService{ShouldFail: false} // Успешный сценарий
	// Для теста компенсации: kitchenStep := &KitchenService{ShouldFail: true} 

	// 5. Сборка и запуск Оркестратора
	steps := []SagaStep{orderStep, paymentStep, kitchenStep}
	orchestrator := NewSagaOrchestrator(steps, sagaRepo)

	// Генерируем ID для нового заказа
	orderID := uuid.New().String()
	log.Printf("ОРКЕСТРАТОР: Запуск саги для нового заказа с ID: %s", orderID)

	// В реальной системе это может делать API-шлюз.
	// Мы создаем "черновик" заказа, который Сага должна будет подтвердить.
	_, err = dbpool.Exec(ctx, "INSERT INTO orders (id, status, price, items) VALUES ($1, 'PENDING', 99.99, '{\"item\": \"pizza\"}')", orderID)
	if err != nil {
		log.Fatalf("ОРКЕСТРАТОР: Не удалось создать начальную запись о заказе: %v", err)
	}

	// Запускаем сагу с таймаутом
	sagaCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := orchestrator.ExecuteSaga(sagaCtx, orderID); err != nil {
		log.Printf("ОРКЕСТРАТОР: Сага завершилась с ошибкой: %v", err)
	} else {
		log.Printf("ОРКЕСТРАТОР: Сага для заказа %s успешно завершена.", orderID)
	}
}
