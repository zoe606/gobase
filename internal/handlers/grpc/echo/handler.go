package echohandler

import (
	"context"

	echov1 "go-boilerplate/pkg/api/echo/v1"
)

type UseCase interface {
	Echo(context.Context, string) (string, error)
}

type Handler struct {
	echov1.UnimplementedEchoServiceServer
	usecase UseCase
}

func New(usecase UseCase) *Handler { return &Handler{usecase: usecase} }
