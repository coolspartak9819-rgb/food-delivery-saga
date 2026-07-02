package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "saga/proto"
)

type server struct {
	pb.UnimplementedKitchenServiceServer
}

func (s *server) ReserveIngredients(ctx context.Context, req *pb.ReserveIngredientsRequest) (*pb.ReserveIngredientsResponse, error) {
	log.Printf("KitchenService: ReserveIngredients-запрос для заказа %s", req.OrderId)

	// Симулируем ошибку, если установлена переменная окружения
	if os.Getenv("KITCHEN_SHOULD_FAIL") == "true" {
		log.Println("KitchenService: Симуляция сбоя - ингредиенты закончились!")
		return nil, status.Error(codes.FailedPrecondition, "Ингредиенты закончились")
	}

	return &pb.ReserveIngredientsResponse{Status: "INGREDIENTS_RESERVED"}, nil
}

func (s *server) ReleaseIngredients(ctx context.Context, req *pb.ReleaseIngredientsRequest) (*pb.ReleaseIngredientsResponse, error) {
	log.Printf("KitchenService: ReleaseIngredients-запрос для заказа %s", req.OrderId)
	return &pb.ReleaseIngredientsResponse{Status: "INGREDIENTS_RELEASED"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterKitchenServiceServer(s, &server{})

	log.Println("KitchenService запущен на порту :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
