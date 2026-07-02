package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "saga/proto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()
	log.Println("ОРКЕСТРАТОР: Успешное подключение к PostgreSQL.")

	// Создаем gRPC соединения
	orderConn := createGRPCConn(os.Getenv("ORDER_SERVICE_ADDR"))
	defer orderConn.Close()
	paymentConn := createGRPCConn(os.Getenv("PAYMENT_SERVICE_ADDR"))
	defer paymentConn.Close()
	kitchenConn := createGRPCConn(os.Getenv("KITCHEN_SERVICE_ADDR"))
	defer kitchenConn.Close()

	// Создаем шаги саги
	steps := []SagaStep{
		&OrderClient{client: pb.NewOrderServiceClient(orderConn)},
		&PaymentClient{client: pb.NewPaymentServiceClient(paymentConn)},
		&KitchenClient{client: pb.NewKitchenServiceClient(kitchenConn)},
	}

	sagaRepo := NewPostgresSagaRepository(dbpool)
	orchestrator := NewSagaOrchestrator(steps, sagaRepo)

	// Запускаем сагу
	orderID := uuid.New().String()
	log.Printf("ОРКЕСТРАТОР: Запуск саги для нового заказа %s", orderID)

	_, err = dbpool.Exec(context.Background(), "INSERT INTO orders (id, status, price, items) VALUES ($1, 'PENDING', 123.45, '{}')", orderID)
	if err != nil {
		log.Fatalf("Failed to create initial order: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := orchestrator.ExecuteSaga(ctx, orderID); err != nil {
		log.Printf("ОРКЕСТРАТОР: Сага завершилась с ошибкой: %v", err)
	}
}
