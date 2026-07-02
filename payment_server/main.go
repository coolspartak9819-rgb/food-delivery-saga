package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "saga/proto"
)

type server struct {
	pb.UnimplementedPaymentServiceServer
}

func (s *server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	log.Printf("PaymentService: ProcessPayment-запрос для заказа %s", req.OrderId)
	// Симуляция успешной оплаты
	return &pb.ProcessPaymentResponse{Status: "PAYMENT_PROCESSED"}, nil
}

func (s *server) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	log.Printf("PaymentService: RefundPayment-запрос для заказа %s", req.OrderId)
	// Симуляция возврата
	return &pb.RefundPaymentResponse{Status: "PAYMENT_REFUNDED"}, nil
}

func main() {
	// Этот сервис для примера не требует БД, но в реальности мог бы
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server{})

	log.Println("PaymentService запущен на порту :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
