package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/1oneday2/saga/proto"
)

// PaymentDB определяет интерфейс для будущих операций с БД.
type PaymentDB interface {
	// Например, в будущем здесь может быть:
	// RecordTransaction(ctx context.Context, orderID string, amount float64) error
}

// mockDB — это временная заглушка, пока реальная БД не нужна.
type mockDB struct{}

// server теперь зависит от интерфейса PaymentDB.
type server struct {
	pb.UnimplementedPaymentServiceServer
	db PaymentDB
}

func (s *server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	log.Printf("PaymentService: ProcessPayment-запрос для заказа %s", req.OrderId)
	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ID заказа не может быть пустым")
	}
	// Здесь могла бы быть логика вызова s.db.RecordTransaction(...)
	return &pb.ProcessPaymentResponse{Status: "PAYMENT_PROCESSED"}, nil
}

func (s *server) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	log.Printf("PaymentService: RefundPayment-запрос для заказа %s", req.OrderId)
	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ID заказа не может быть пустым")
	}
	return &pb.RefundPaymentResponse{Status: "PAYMENT_REFUNDED"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Внедряем нашу заглушку в сервер.
	db := &mockDB{}
	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server{db: db})

	log.Println("PaymentService запущен на порту :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
