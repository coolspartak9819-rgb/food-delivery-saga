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

	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	pb.UnimplementedOrderServiceServer
	db *pgxpool.Pool
}

func (s *server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log.Printf("OrderService: CreateOrder-запрос для заказа %s", req.OrderId)
	_, err := s.db.Exec(ctx, "UPDATE orders SET status = 'CREATED' WHERE id = $1", req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DB error: %v", err)
	}
	return &pb.CreateOrderResponse{Status: "ORDER_CREATED"}, nil
}

func (s *server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	log.Printf("OrderService: CancelOrder-запрос для заказа %s", req.OrderId)
	_, err := s.db.Exec(ctx, "UPDATE orders SET status = 'CANCELLED' WHERE id = $1", req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DB error: %v", err)
	}
	return &pb.CancelOrderResponse{Status: "ORDER_CANCELLED"}, nil
}

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

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, &server{db: dbpool})

	log.Println("OrderService запущен на порту :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
