package auth

import (
	"auth/internal/model"
	"context"
)

type SessUseCase interface {
	CreateSession(ctx context.Context, email string, password string) (*model.LoginResponse, error)
}
