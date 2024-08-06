package delivery

import (
	"auth/internal/auth"
	sessionService "auth/pb/sessions"
	"auth/pkg/logger"
	"context"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (s *SessionGRPCService) CreateSession(ctx context.Context, req *sessionService.LoginUserRequest) (*sessionService.LoginUserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userServiceGrpc.Register")
	defer span.Finish()
	res, err := s.sessUseCase.CreateSession(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &sessionService.LoginUserResponse{
		SessionId:             res.SessionID,
		AccessToken:           res.AccessToken,
		RefreshToken:          res.RefreshToken,
		AccessTokenExpiresAt:  timestamppb.New(res.AccessTokenExpiresAt),
		RefreshTokenExpiresAt: timestamppb.New(res.RefreshTokenExpiresAt),
	}, nil
}
