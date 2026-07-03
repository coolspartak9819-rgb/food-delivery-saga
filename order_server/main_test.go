package main

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	pb "github.com/1oneday2/saga/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// mockOrderDB — это наша "заглушка" для базы данных.
type mockOrderDB struct {
	shouldFail bool
}

func (m *mockOrderDB) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	if m.shouldFail {
		return errors.New("симуляция ошибки базы данных")
	}
	log.Printf("mockDB: заказ %s успешно обновлен на статус %s", orderID, status)
	return nil
}

const bufSize = 1024 * 1024

func setup(t *testing.T, dbMock OrderDB) (pb.OrderServiceClient, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, &server{db: dbMock})

	go func() {
		if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("Сервер завершился с ошибкой: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Не удалось подключиться к bufnet: %v", err)
	}

	client := pb.NewOrderServiceClient(conn)

	cleanup := func() {
		conn.Close()
		s.GracefulStop()
	}

	return client, cleanup
}

func TestCreateOrder(t *testing.T) {
	testCases := []struct {
		name         string
		req          *pb.CreateOrderRequest
		dbMock       *mockOrderDB
		expectedCode codes.Code
	}{
		{
			name:         "Успешное создание заказа",
			req:          &pb.CreateOrderRequest{OrderId: "valid-id-123"},
			dbMock:       &mockOrderDB{shouldFail: false},
			expectedCode: codes.OK,
		},
		{
			name:         "Ошибка валидации - пустой ID",
			req:          &pb.CreateOrderRequest{OrderId: ""},
			dbMock:       &mockOrderDB{},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Ошибка базы данных при создании",
			req:          &pb.CreateOrderRequest{OrderId: "db-fail-id"},
			dbMock:       &mockOrderDB{shouldFail: true},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, cleanup := setup(t, tc.dbMock)
			defer cleanup()

			_, err := client.CreateOrder(context.Background(), tc.req)

			st, ok := status.FromError(err)
			if !ok {
				st = status.New(codes.OK, "")
			}

			if st.Code() != tc.expectedCode {
				t.Errorf("Ожидался код ошибки '%s', но получен '%s'. Полная ошибка: %v", tc.expectedCode, st.Code(), err)
			}
		})
	}
}
