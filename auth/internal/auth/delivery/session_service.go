package delivery

import (
	"auth/internal/auth"
	sessionService "auth/pb/sessions"
	"auth/pkg/logger"
)

type SessionGRPCService struct {
	sessionService.UnimplementedAuthServiceServer
	sessUseCase auth.SessUseCase
	log         logger.Logger
}

func NewSessionGRPCService(sessUseCase auth.SessUseCase, log logger.Logger) *SessionGRPCService {
	return &SessionGRPCService{
		sessUseCase: sessUseCase,
		log:         log,
	}
}
