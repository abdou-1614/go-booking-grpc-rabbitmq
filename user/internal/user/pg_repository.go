package user

import (
	"Go-grpc/internal/model"
	"context"

	uuid "github.com/satori/go.uuid"
)

type PGRepository interface {
	Create(ctx context.Context, user *model.User) (*model.UserResponse, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error)
	Update(ctx context.Context, user *model.UserUpdate) (*model.UserResponse, error)
	UpdateAvatar(ctx context.Context, data *model.UpdateAvatarMsg) error
	GetUsersByIDs(ctx context.Context, userIDs []string) ([]*model.UserResponse, error)
}
