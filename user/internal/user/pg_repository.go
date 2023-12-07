package user

import (
	"Go-grpc/internal/model"
	"context"

	uuid "github.com/satori/go.uuid"
)

type PGRepository interface {
	Create(ctx context.Context, user *model.User) (*model.UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.UserResponse, error)
	UpdateAvatar(ctx context.Context, data model.UploadedImageMsg) (*model.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*model.UserResponse, error)
}
