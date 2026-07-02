// order_client.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	pb "saga/order"
)

// OrderServiceGRPCClient - это адаптер, который реализует интерфейс SagaStep
// и инкапсулирует логику gRPC-вызовов к удаленному OrderService.
type OrderServiceGRPCClient struct {
	client pb.OrderServiceClient
}

// NewOrderServiceGRPCClient создает новый gRPC-клиент для сервиса заказов.
func NewOrderServiceGRPCClient(conn *grpc.ClientConn) *OrderServiceGRPCClient {
	return &OrderServiceGRPCClient{
		client: pb.NewOrderServiceClient(conn),
	}
}

func (c *OrderServiceGRPCClient) Name() string {
	return "OrderService(gRPC)"
}

// Execute вызывает удаленный метод CreateOrder через gRPC.
func (c *OrderServiceGRPCClient) Execute(ctx context.Context, orderID string) error {
	log.Printf("[%s] Отправка gRPC-запроса CreateOrder для заказа %s", c.Name(), orderID)

	// Устанавливаем таймаут для gRPC-вызова
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.CreateOrderRequest{
		OrderId:   orderID,
		Price:     99.99, // В реальной системе цена придет из другого источника
		ItemsJson: `{"item": "pizza", "quantity": 1}`,
	}

	_, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		// Корректно обрабатываем gRPC-ошибки
		st, ok := status.FromError(err)
		if ok {
			// Это структурированная ошибка от gRPC-сервера
			log.Printf("Ошибка от gRPC-сервера OrderService: code=%s, message=%s", st.Code(), st.Message())
			return fmt.Errorf("gRPC-ошибка при создании заказа: %s", st.Message())
		}
		// Это может быть ошибка сети или другая проблема с соединением
		return fmt.Errorf("неизвестная ошибка при вызове CreateOrder: %w", err)
	}

	log.Printf("[%s] Заказ %s успешно создан через gRPC", c.Name(), orderID)
	return nil
}

// Compensate вызывает удаленный метод CancelOrder через gRPC.
func (c *OrderServiceGRPCClient) Compensate(ctx context.Context, orderID string) error {
	log.Printf("[%s] Отправка gRPC-запроса CancelOrder для заказа %s", c.Name(), orderID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.CancelOrderRequest{OrderId: orderID}

	_, err := c.client.CancelOrder(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("Ошибка от gRPC-сервера OrderService при компенсации: code=%s, message=%s", st.Code(), st.Message())
			return fmt.Errorf("gRPC-ошибка при отмене заказа: %s", st.Message())
		}
		return fmt.Errorf("неизвестная ошибка при вызове CancelOrder: %w", err)
	}

	log.Printf("[%s] Заказ %s успешно отменен через gRPC", c.Name(), orderID)
	return nil
}
