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
	uuid "github.com/satori/go.uuid"
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

func (u *UserGRPCService) GetUserID(ctx context.Context, req *userService.GetByIDRequest) (*userService.GetByIDResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserGRPCService.CreateUser")

	defer span.Finish()

	userUUID, err := uuid.FromString(req.GetUserID())

	if err != nil {
		u.logger.Errorf("uuid.FromString: %v", err)
		return nil, grpc_error.ErrorResponse(err, "uuid.FromString")
	}

	foundUser, err := u.userUC.GetByID(ctx, userUUID)

	if err != nil {
		u.logger.Errorf("userUC.GetByID : %v", err)
		return nil, grpc_error.ErrorResponse(err, "userUC.GetByID")
	}

	return &userService.GetByIDResponse{User: foundUser.ToProto()}, nil
}

func (u *UserGRPCService) CreateUser(ctx context.Context, req *userService.CreateUserRequest) (*userService.CreateUserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserGRPCService.CreateUser")

	defer span.Finish()

	user := &model.User{
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
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
