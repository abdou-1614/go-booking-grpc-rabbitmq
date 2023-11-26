package user

import (
	"Go-grpc/internal/model"
	"context"

	uuid "github.com/satori/go.uuid"
)

type UseCase interface {
	Register(ctx context.Context, user *model.User) (*model.UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.UserResponse, error)
}
