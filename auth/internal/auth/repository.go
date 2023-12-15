package auth

import (
	"auth/internal/model"
	"context"
)

type Repository interface {
	CreateSession(ctx context.Context, arg *model.CreateSessionParams) (*model.SessionResponse, error)
}
