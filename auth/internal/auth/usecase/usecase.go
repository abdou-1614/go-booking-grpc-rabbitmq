package usecase

import (
	"Go-grpc/pkg/logger"
	"auth/config"
	"auth/internal/auth"
	"auth/internal/auth/token"
	"auth/internal/interceptors"
	"auth/internal/model"
	userService "auth/pb/user"
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type SessionUseCase struct {
	sessionPGRepo auth.Repository
	log           logger.Loggor
	userClient    userService.UserServiceClient
	cfg           *config.Config
	tokenMake     token.Maker
	tracer        opentracing.Tracer
}

func NewSessionUseCase(sessionPGRepo auth.Repository, log logger.Loggor, userClient userService.UserServiceClient, cfg *config.Config, tokenMaker token.Maker, tracer opentracing.Tracer) *SessionUseCase {
	return &SessionUseCase{
		sessionPGRepo: sessionPGRepo,
		log:           log,
		userClient:    userClient,
		cfg:           cfg,
		tokenMake:     tokenMaker,
		tracer:        tracer,
	}
}

func (s *SessionUseCase) CreateSession(ctx context.Context, email string, password string) (*model.LoginResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionUseCase.CreateSession")

	defer span.Finish()

	md := interceptors.NewInterceptorManager(s.log, s.cfg, s.tracer)

	userResponse, err := s.userClient.GetUserEmail(ctx, &userService.GetByEmailRequest{Email: email})

	s.log.Infof("User_Email ==> %v, User_Password ==> %v", userResponse.User.Email, userResponse.User.Password)

	if err != nil {
		return nil, errors.Wrap(err, "userResponse.GetEmail")
	}

	err = bcrypt.CompareHashAndPassword([]byte(userResponse.User.Password), []byte(password))

	if err != nil {
		return nil, errors.Wrap(err, "Invalid Password")
	}

	accessToken, accessPayload, err := s.tokenMake.CreateToken(userResponse.User.Email, userResponse.User.Role, time.Duration(s.cfg.GRPCServer.AccessTokenExpire))

	if err != nil {
		return nil, errors.Wrap(err, "failed to create access token")
	}

	refreshToken, refreshPayload, err := s.tokenMake.CreateToken(userResponse.User.Email, userResponse.User.Role, time.Duration(s.cfg.GRPCServer.RefreshTokenExpire))

	if err != nil {
		return nil, errors.Wrap(err, "failed to create Refresh token")
	}

	mtdt := md.ExtractMetadata(ctx)

	session, err := s.sessionPGRepo.CreateSession(ctx, &model.CreateSessionParams{
		ID:           uuid.UUID(refreshPayload.ID),
		Email:        userResponse.User.Email,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIP:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	return &model.LoginResponse{
		SessionID:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}, nil
}
