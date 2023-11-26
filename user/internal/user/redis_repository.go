package user

import (
	"Go-grpc/internal/model"
	"context"

	uuid "github.com/satori/go.uuid"
)

type RedisRepository interface {
	SaveUser(ctx context.Context, user *model.UserResponse) error
	GetUserID(ctx context.Context, id uuid.UUID) (*model.UserResponse, error)
}
