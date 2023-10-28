package user

import (
	"Go-grpc/internal/model"
	"context"
)

type RedisRepository interface {
	SaveUser(ctx context.Context, user *model.UserResponse) error
}
