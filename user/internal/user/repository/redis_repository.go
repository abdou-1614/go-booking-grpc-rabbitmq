package repository

import (
	"Go-grpc/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
)

type userRedisRepository struct {
	redisConn  *redis.Client
	prefix     string
	expiration time.Duration
}

func NewUserRedisRepository(redisConn *redis.Client, prefix string, expiration time.Duration) *userRedisRepository {
	return &userRedisRepository{redisConn: redisConn, prefix: prefix, expiration: expiration}
}

func (u *userRedisRepository) SaveUser(ctx context.Context, user *model.UserResponse) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRedisRepository")

	defer span.Finish()

	userBytes, err := json.Marshal(user)

	if err != nil {
		return errors.Wrap(err, "userRedisRepository.SaveUser.json.Marshal")
	}

	if err := u.redisConn.SetEx(ctx, u.createKey(user.ID), string(userBytes), u.expiration).Err(); err != nil {
		return errors.Wrap(err, "userRedisRepository.SaveUser.json.Marshal")
	}

	return nil
}

func (u *userRedisRepository) createKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s: %s", u.prefix, userID)
}
