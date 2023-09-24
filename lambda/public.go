package lambda

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/ihippik/lambda-go/lambda/proto"
)

// Handler is a user function that handles lambda requests.
type Handler func(ctx context.Context, payload []byte) ([]byte, error)

// Server is a wrapper for user Handler.
type Server struct {
	proto.UnimplementedLambdaServerServer
	handler Handler
}

// Start starts the lambda handler.
func Start(handler Handler) {
	const serverAddr = ":8080"

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		slog.Error(err.Error())
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()

	proto.RegisterLambdaServerServer(grpcServer, &Server{handler: handler})

	if err := grpcServer.Serve(lis); err != nil {
		slog.Error(err.Error())
	}
}

func (h *Server) MakeRequest(ctx context.Context, payload *proto.Payload) (*proto.Payload, error) {
	slog.Debug("got request", "payload_size", len(payload.Data))

	respData, err := h.handler(ctx, payload.Data)
	if err != nil {
		return nil, fmt.Errorf("handler: %w", err)
	}

	slog.Debug("got response", "payload_size", "response_size", len(respData))

	return &proto.Payload{Data: respData}, nil
}
