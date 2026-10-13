package echohandler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	echousecase "go-boilerplate/internal/usecase/echo"
	echov1 "go-boilerplate/pkg/api/echo/v1"
	"go-boilerplate/pkg/grpcserver"
)

func (h *Handler) Echo(ctx context.Context, request *echov1.EchoRequest) (*echov1.EchoResponse, error) {
	message, err := h.usecase.Echo(ctx, request.GetMessage())
	if err != nil {
		switch {
		case errors.Is(err, echousecase.ErrInvalidMessage):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, status.FromContextError(err).Err()
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return &echov1.EchoResponse{Message: message, RequestId: grpcserver.RequestID(ctx)}, nil
}
