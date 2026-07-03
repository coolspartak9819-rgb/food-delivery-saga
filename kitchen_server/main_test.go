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

// mockKitchenDB для тестов.
type mockKitchenDB struct{}

const bufSize = 1024 * 1024

func setup(t *testing.T, dbMock KitchenDB) (pb.KitchenServiceClient, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterKitchenServiceServer(s, &server{db: dbMock})

	go func() {
		if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	cleanup := func() {
		conn.Close()
		s.GracefulStop()
	}

	return pb.NewKitchenServiceClient(conn), cleanup
}

func TestKitchenService(t *testing.T) {
	dbMock := &mockKitchenDB{}

	t.Run("ReserveIngredients_Success", func(t *testing.T) {
		client, cleanup := setup(t, dbMock)
		defer cleanup()

		_, err := client.ReserveIngredients(context.Background(), &pb.ReserveIngredientsRequest{OrderId: "valid-order-id"})
		if err != nil {
			t.Fatalf("Ожидался успех, но получена ошибка: %v", err)
		}
	})

	t.Run("ReserveIngredients_SimulatedFailure", func(t *testing.T) {
		// Устанавливаем переменную окружения для симуляции сбоя
		t.Setenv("KITCHEN_SHOULD_FAIL", "true")
		
		client, cleanup := setup(t, dbMock)
		defer cleanup()

		_, err := client.ReserveIngredients(context.Background(), &pb.ReserveIngredientsRequest{OrderId: "fail-id"})
		st, _ := status.FromError(err)
		if st.Code() != codes.FailedPrecondition {
			t.Errorf("Ожидался код %s, но получен %s", codes.FailedPrecondition, st.Code())
		}
	})

	t.Run("ReserveIngredients_InvalidID", func(t *testing.T) {
		client, cleanup := setup(t, dbMock)
		defer cleanup()

		_, err := client.ReserveIngredients(context.Background(), &pb.ReserveIngredientsRequest{OrderId: ""})
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("Ожидался код %s, но получен %s", codes.InvalidArgument, st.Code())
		}
	})
}
