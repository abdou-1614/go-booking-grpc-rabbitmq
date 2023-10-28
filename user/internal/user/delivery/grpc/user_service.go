package grpc

import (
	"Go-grpc/internal/model"
	"Go-grpc/internal/user"
	userService "Go-grpc/pb"
	"Go-grpc/pkg/grpc_error"
	"Go-grpc/pkg/logger"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type UserGRPCService struct {
	userService.UnimplementedUserServiceServer
	userUC   user.UseCase
	logger   logger.Loggor
	validate *validator.Validate
}

func NewUserGRPCService(userUC user.UseCase, logger logger.Loggor, validate *validator.Validate) *UserGRPCService {
	return &UserGRPCService{
		userUC:   userUC,
		logger:   logger,
		validate: validate,
	}
}

func (u *UserGRPCService) CreateUser(ctx context.Context, req *userService.CreateUserRequest) (*userService.CreateUserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserGRPCService.CreateUser")

	defer span.Finish()

	var userModel *model.User

	if err := userModel.PrepareToCreate(); err != nil {
		return nil, errors.Wrap(err, "UserGRPCService.PrepareToCreate")
	}

	role := model.Role(req.GetRole())

	user := &model.User{
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		Role:      &role,
	}

	if err := u.validate.StructCtx(ctx, user); err != nil {
		return nil, grpc_error.ErrorResponse(err, err.Error())
	}

	createdUser, err := u.userUC.Register(ctx, user)

	if err != nil {
		u.logger.Errorf("userUC.CreateUser : %v", err)
		return nil, grpc_error.ErrorResponse(err, "userUC.CreateUser")
	}

	return &userService.CreateUserResponse{User: createdUser.ToProto()}, nil
}
