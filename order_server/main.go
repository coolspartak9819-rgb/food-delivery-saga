// order_server/main.go
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Импортируем сгенерированный код из нашего же модуля
	pb "saga/order"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	port               = ":50051"
	dbConnectionString = "postgres://user:password@localhost:5432/saga_db?sslmode=disable"
)

// server реализует сгенерированный интерфейс OrderServiceServer.
type server struct {
	pb.UnimplementedOrderServiceServer
	db *pgxpool.Pool
}

// CreateOrder - реализация gRPC-метода для создания заказа.
func (s *server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log.Printf("Получен gRPC-запрос на создание заказа: ID %s", req.OrderId)

	// В реальной системе мы бы создавали заказ в транзакции.
	// Здесь мы просто обновляем статус существующей записи, созданной оркестратором.
	_, err := s.db.Exec(ctx, "UPDATE orders SET status = 'CREATED' WHERE id = $1", req.OrderId)
	if err != nil {
		log.Printf("Ошибка БД при создании заказа %s: %v", req.OrderId, err)
		// Возвращаем корректную gRPC-ошибку
		return nil, status.Errorf(codes.Internal, "не удалось сохранить заказ в БД: %v", err)
	}

	log.Printf("Заказ %s успешно создан в БД", req.OrderId)
	return &pb.CreateOrderResponse{Status: "ORDER_CREATED"}, nil
}

// CancelOrder - реализация gRPC-метода для отмены заказа (компенсация).
func (s *server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	log.Printf("Получен gRPC-запрос на отмену заказа: ID %s", req.OrderId)

	_, err := s.db.Exec(ctx, "UPDATE orders SET status = 'CANCELLED' WHERE id = $1", req.OrderId)
	if err != nil {
		log.Printf("Ошибка БД при отмене заказа %s: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "не удалось отменить заказ в БД: %v", err)
	}

	log.Printf("Заказ %s успешно отменен в БД", req.OrderId)
	return &pb.CancelOrderResponse{Status: "ORDER_CANCELLED"}, nil
}

func main() {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, dbConnectionString)
	if err != nil {
		log.Fatalf("Не удалось создать пул соединений с БД: %v", err)
	}
	defer dbpool.Close()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Не удалось прослушать порт %s: %v", port, err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, &server{db: dbpool})

	log.Printf("gRPC-сервер OrderService запущен и слушает порт %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
