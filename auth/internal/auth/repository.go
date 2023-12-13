package auth

import (
	"auth/internal/model"
	"context"
)

type Repository interface {
	Login(ctx context.Context, arg *model.CreateSessionParams) (*model.SessionResponse, error)
}
