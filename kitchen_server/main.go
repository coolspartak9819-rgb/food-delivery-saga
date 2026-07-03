package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/1oneday2/saga/proto"
)

// KitchenDB определяет интерфейс для будущих операций с БД.
type KitchenDB interface {
	// Например, в будущем здесь может быть:
	// CheckAndReserve(ctx context.Context, orderID string, items []string) error
}

// mockDB — это временная заглушка.
type mockDB struct{}

// server теперь зависит от интерфейса KitchenDB.
type server struct {
	pb.UnimplementedKitchenServiceServer
	db KitchenDB
}

func (s *server) ReserveIngredients(ctx context.Context, req *pb.ReserveIngredientsRequest) (*pb.ReserveIngredientsResponse, error) {
	log.Printf("KitchenService: ReserveIngredients-запрос для заказа %s", req.OrderId)
	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ID заказа не может быть пустым")
	}

	// Симулируем ошибку, если установлена переменная окружения
	if os.Getenv("KITCHEN_SHOULD_FAIL") == "true" {
		log.Println("KitchenService: Симуляция сбоя - ингредиенты закончились!")
		return nil, status.Error(codes.FailedPrecondition, "Ингредиенты закончились")
	}

	return &pb.ReserveIngredientsResponse{Status: "INGREDIENTS_RESERVED"}, nil
}

func (s *server) ReleaseIngredients(ctx context.Context, req *pb.ReleaseIngredientsRequest) (*pb.ReleaseIngredientsResponse, error) {
	log.Printf("KitchenService: ReleaseIngredients-запрос для заказа %s", req.OrderId)
	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ID заказа не может быть пустым")
	}
	return &pb.ReleaseIngredientsResponse{Status: "INGREDIENTS_RELEASED"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	db := &mockDB{}
	s := grpc.NewServer()
	pb.RegisterKitchenServiceServer(s, &server{db: db})

	log.Println("KitchenService запущен на порту :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
