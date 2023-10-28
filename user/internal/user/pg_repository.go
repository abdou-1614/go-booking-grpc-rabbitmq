package user

import (
	"Go-grpc/internal/model"
	"context"
)

type PGRepository interface {
	Create(ctx context.Context, user *model.User) (*model.UserResponse, error)
}
