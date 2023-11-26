package usecase

import (
	"Go-grpc/internal/model"
	"Go-grpc/internal/user"
	"Go-grpc/internal/user/delivery/rabbitmq"
	"Go-grpc/pkg/logger"
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

const (
	imagesExchange = "images"
	resizeKey      = "resize_image_key"
	userUUIDHeader = "user_uuid"
)

type userUseCase struct {
	userPGRepo user.PGRepository
	log        logger.Loggor
	redRepo    user.RedisRepository
	amqp       rabbitmq.Publisher
}

func NewUserUseCase(userPGRepo user.PGRepository, log logger.Loggor, redRepo user.RedisRepository, amqp rabbitmq.Publisher) *userUseCase {
	return &userUseCase{
		userPGRepo: userPGRepo,
		log:        log,
		redRepo:    redRepo,
		amqp:       amqp,
	}
}

func (u *userUseCase) Register(ctx context.Context, user *model.User) (*model.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.Register")
	defer span.Finish()

	if err := user.PrepareToCreate(); err != nil {
		return nil, errors.Wrap(err, "user.PrepareCreate")
	}

	created, err := u.userPGRepo.Create(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "userPGRepo.Create")
	}

	return created, err
}

func (u *userUseCase) GetByID(ctx context.Context, id uuid.UUID) (*model.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.GetByID")

	defer span.Finish()

	cashedUser, err := u.redRepo.GetUserID(ctx, id)

	if err != nil {
		u.log.Errorf("redisRepo.GetUserID : %v", err)
	}

	if cashedUser != nil {
		return cashedUser, nil
	}

	userResponse, err := u.userPGRepo.GetByID(ctx, id)

	if err != nil {
		return nil, errors.Wrap(err, "userUseCase.userPGRepo.GetByID")
	}

	if err := u.redRepo.SaveUser(ctx, userResponse); err != nil {
		u.log.Errorf("redisRepo.SaveUser: %v", err)
	}

	return userResponse, nil
}
