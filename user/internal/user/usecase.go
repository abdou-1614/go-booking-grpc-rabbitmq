package user

import (
	"Go-grpc/internal/model"
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type UseCase interface {
	Register(ctx context.Context, user *model.User) (*model.UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.UserResponse, error)
	UpdateUploadedAvatar(ctx context.Context, delevery amqp.Delivery) error
	UpdateAvatar(ctx context.Context, data *model.UpdateAvatarMsg) error
}
