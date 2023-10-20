package user

import (
	"Go-grpc/internal/model"
	"context"

	uuid "github.com/satori/go.uuid"
)

type RedisRepository interface {
	SaveUser(ctx context.Context, user *model.UserResponse) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}
