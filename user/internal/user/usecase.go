package user

import (
	"Go-grpc/internal/model"
	"context"
)

type UseCase interface {
	Register(ctx context.Context, user *model.User) (*model.UserResponse, error)
}
