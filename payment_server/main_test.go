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

// mockPaymentDB для тестов.
type mockPaymentDB struct{}

const bufSize = 1024 * 1024

func setup(t *testing.T, dbMock PaymentDB) (pb.PaymentServiceClient, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server{db: dbMock})

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

	return pb.NewPaymentServiceClient(conn), cleanup
}

func TestPaymentService(t *testing.T) {
	dbMock := &mockPaymentDB{}
	client, cleanup := setup(t, dbMock)
	defer cleanup()

	t.Run("ProcessPayment_Success", func(t *testing.T) {
		_, err := client.ProcessPayment(context.Background(), &pb.ProcessPaymentRequest{OrderId: "valid-order-id"})
		if err != nil {
			t.Fatalf("Ожидался успех, но получена ошибка: %v", err)
		}
	})

	t.Run("ProcessPayment_InvalidID", func(t *testing.T) {
		_, err := client.ProcessPayment(context.Background(), &pb.ProcessPaymentRequest{OrderId: ""})
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("Ожидался код %s, но получен %s", codes.InvalidArgument, st.Code())
		}
	})

	t.Run("RefundPayment_Success", func(t *testing.T) {
		_, err := client.RefundPayment(context.Background(), &pb.RefundPaymentRequest{OrderId: "valid-order-id"})
		if err != nil {
			t.Fatalf("Ожидался успех, но получена ошибка: %v", err)
		}
	})

	t.Run("RefundPayment_InvalidID", func(t *testing.T) {
		_, err := client.RefundPayment(context.Background(), &pb.RefundPaymentRequest{OrderId: ""})
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			t.Errorf("Ожидался код %s, но получен %s", codes.InvalidArgument, st.Code())
		}
	})
}
