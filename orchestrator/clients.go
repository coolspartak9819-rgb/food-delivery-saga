package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "saga/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// --- Order Service Client ---

type OrderClient struct {
	client pb.OrderServiceClient
}

func (c *OrderClient) Name() string { return "OrderService" }

func (c *OrderClient) Execute(ctx context.Context, orderID string) error {
	_, err := c.client.CreateOrder(ctx, &pb.CreateOrderRequest{OrderId: orderID})
	return handleGRPCError(err, c.Name(), "Execute")
}

func (c *OrderClient) Compensate(ctx context.Context, orderID string) error {
	_, err := c.client.CancelOrder(ctx, &pb.CancelOrderRequest{OrderId: orderID})
	return handleGRPCError(err, c.Name(), "Compensate")
}

// --- Payment Service Client ---

type PaymentClient struct {
	client pb.PaymentServiceClient
}

func (c *PaymentClient) Name() string { return "PaymentService" }

func (c *PaymentClient) Execute(ctx context.Context, orderID string) error {
	_, err := c.client.ProcessPayment(ctx, &pb.ProcessPaymentRequest{OrderId: orderID})
	return handleGRPCError(err, c.Name(), "Execute")
}

func (c *PaymentClient) Compensate(ctx context.Context, orderID string) error {
	_, err := c.client.RefundPayment(ctx, &pb.RefundPaymentRequest{OrderId: orderID})
	return handleGRPCError(err, c.Name(), "Compensate")
}

// --- Kitchen Service Client ---

type KitchenClient struct {
	client pb.KitchenServiceClient
}

func (c *KitchenClient) Name() string { return "KitchenService" }

func (c *KitchenClient) Execute(ctx context.Context, orderID string) error {
	_, err := c.client.ReserveIngredients(ctx, &pb.ReserveIngredientsRequest{OrderId: orderID})
	return handleGRPCError(err, c.Name(), "Execute")
}

func (c *KitchenClient) Compensate(ctx context.Context, orderID string) error {
	_, err := c.client.ReleaseIngredients(ctx, &pb.ReleaseIngredientsRequest{OrderId: orderID})
	return handleGRPCError(err, c.Name(), "Compensate")
}

// --- Helper Functions ---

func handleGRPCError(err error, serviceName, methodName string) error {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("gRPC error from %s.%s: %s (%s)", serviceName, methodName, st.Message(), st.Code())
		}
		return fmt.Errorf("unknown gRPC error from %s.%s: %w", serviceName, methodName, err)
	}
	return nil
}

func createGRPCConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", addr, err)
	}
	return conn
}
