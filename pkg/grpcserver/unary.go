package grpcserver

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type requestIDKey struct{}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func (s *Server) unary(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
	started := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	id := uuid.NewString()
	if values := md.Get("x-request-id"); len(values) == 1 && values[0] != "" && len(values[0]) <= 128 {
		id = values[0]
	}
	ctx = context.WithValue(ctx, requestIDKey{}, id)
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", id))
	defer func() {
		if recovered := recover(); recovered != nil {
			s.l.Error("gRPC panic method=%s request_id=%s: %v", info.FullMethod, id, recovered)
			response, err = nil, status.Error(codes.Internal, "internal error")
		}
		s.l.Info("gRPC method=%s request_id=%s code=%s duration=%s", info.FullMethod, id, status.Code(err), time.Since(started))
	}()
	deadline := time.Now().Add(s.cfg.RequestTimeout)
	if info.FullMethod != "/grpc.health.v1.Health/Check" {
		expiry, authErr := s.authenticate(md)
		if authErr != nil {
			return nil, authErr
		}
		if expiry.Before(deadline) {
			deadline = expiry
		}
	}
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	response, err = handler(ctx, request)
	if ctx.Err() != nil {
		return nil, status.FromContextError(ctx.Err()).Err()
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, status.FromContextError(err).Err()
	}
	if _, ok := status.FromError(err); !ok {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return response, err
}

func (s *Server) authenticate(md metadata.MD) (time.Time, error) {
	values := md.Get("authorization")
	if len(values) != 1 {
		return time.Time{}, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	parts := strings.Split(values[0], " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return time.Time{}, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	claims, err := s.auth.ValidateToken(parts[1])
	if err != nil || claims == nil || claims.UserID == 0 || claims.ExpiresAt == nil {
		return time.Time{}, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return claims.ExpiresAt.Time, nil
}
