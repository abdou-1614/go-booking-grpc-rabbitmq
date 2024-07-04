package usecase

import (
	"Go-grpc/internal/model"
	"Go-grpc/internal/user"
	"Go-grpc/pkg/http_error"
	"Go-grpc/pkg/logger"
	"context"
	"encoding/json"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
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
}

func NewUserUseCase(userPGRepo user.PGRepository, log logger.Loggor, redRepo user.RedisRepository) *userUseCase {
	return &userUseCase{
		userPGRepo: userPGRepo,
		log:        log,
		redRepo:    redRepo,
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

func (u *userUseCase) GetByEmail(ctx context.Context, email string) (*model.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.GetByEmail")

	defer span.Finish()

	email = strings.ToLower(strings.TrimSpace(email))

	userResponse, err := u.userPGRepo.GetByEmail(ctx, email)

	if err != nil {
		return nil, errors.Wrap(err, "userUseCase.userPGRepo.GetByEmail")
	}

	return userResponse, nil
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

func (u *userUseCase) UpdateAvatar(ctx context.Context, data *model.UpdateAvatarMsg) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.UpdateAvatar")
	defer span.Finish()

	headers := make(amqp.Table, 1)
	headers[userUUIDHeader] = data.UserID.String()

	u.log.Infof("UploadAvatar Publish -%v", headers)
	return nil

}

func (u *userUseCase) UpdateUploadedAvatar(ctx context.Context, delivery amqp.Delivery) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.UpdateUploadedAvatar")

	defer span.Finish()

	var img model.Image

	if err := json.Unmarshal(delivery.Body, &img); err != nil {
		return errors.Wrap(err, "")
	}

	userUUID, ok := delivery.Headers[userUUIDHeader].(string)

	if !ok {
		return errors.Wrap(http_error.InvalidUUID, "delivery.Headers")
	}
	uid, err := uuid.FromString(userUUID)

	if err != nil {
		return errors.Wrap(err, "uuid.FromString")
	}
	created, err := u.userPGRepo.UpdateAvatar(ctx, model.UploadedImageMsg{
		ImageID:    img.ImageID,
		UserID:     uid,
		ImageURL:   img.ImageURL,
		IsUploaded: img.IsUploaded,
	})

	if err != nil {
		return err
	}

	u.log.Infof("UpdateUploadedAvatar", created.ID)
	return nil
}
