package auth

import (
	"auth/internal/model"
	"context"
)

type SessUseCase interface {
	Login(ctx context.Context, arg *model.CreateSessionParams) (*model.SessionResponse, error)
}
