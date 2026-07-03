package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/1oneday2/saga/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrderDB определяет интерфейс для операций с БД, необходимых серверу.
type OrderDB interface {
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
}

// postgresDB — это реальная, production-реализация интерфейса OrderDB.
type postgresDB struct {
	pool *pgxpool.Pool
}

// UpdateOrderStatus выполняет SQL-запрос к настоящей базе данных.
func (p *postgresDB) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	_, err := p.pool.Exec(ctx, "UPDATE orders SET status = $1 WHERE id = $2", status, orderID)
	return err
}

// server теперь зависит от интерфейса OrderDB, а не от конкретной реализации.
type server struct {
	pb.UnimplementedOrderServiceServer
	db OrderDB
}

// CreateOrder теперь содержит логику валидации входных данных.
func (s *server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log.Printf("OrderService: получен CreateOrder-запрос для заказа %s", req.OrderId)

	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ID заказа не может быть пустым")
	}

	if err := s.db.UpdateOrderStatus(ctx, req.OrderId, "CREATED"); err != nil {
		log.Printf("OrderService: ошибка БД при создании заказа %s: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "внутренняя ошибка при работе с БД: %v", err)
	}

	log.Printf("OrderService: заказ %s успешно создан", req.OrderId)
	return &pb.CreateOrderResponse{Status: "ORDER_CREATED"}, nil
}

func (s *server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	log.Printf("OrderService: получен CancelOrder-запрос для заказа %s", req.OrderId)
	if req.OrderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ID заказа не может быть пустым")
	}

	if err := s.db.UpdateOrderStatus(ctx, req.OrderId, "CANCELLED"); err != nil {
		log.Printf("OrderService: ошибка БД при отмене заказа %s: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "внутренняя ошибка при работе с БД: %v", err)
	}

	log.Printf("OrderService: заказ %s успешно отменен", req.OrderId)
	return &pb.CancelOrderResponse{Status: "ORDER_CANCELLED"}, nil
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Переменная окружения DATABASE_URL не установлена")
	}

	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbpool.Close()

	db := &postgresDB{pool: dbpool}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка прослушивания порта: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, &server{db: db})

	log.Println("OrderService запущен на порту :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
