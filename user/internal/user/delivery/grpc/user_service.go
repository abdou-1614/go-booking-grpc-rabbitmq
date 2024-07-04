package grpc

import (
	"Go-grpc/internal/model"
	"Go-grpc/internal/user"
	"Go-grpc/internal/user/delivery/kafka"
	userService "Go-grpc/pb"
	"Go-grpc/pkg/grpc_error"
	"Go-grpc/pkg/logger"
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/go-playground/validator/v10"
	"github.com/opentracing/opentracing-go"
)

type UserGRPCService struct {
	userService.UnimplementedUserServiceServer
	userUC        user.UseCase
	logger        logger.Loggor
	validate      *validator.Validate
	KafkaProducer kafka.UserProducer
}

func NewUserGRPCService(userUC user.UseCase, logger logger.Loggor, validate *validator.Validate, KafkaProducer kafka.UserProducer) *UserGRPCService {
	return &UserGRPCService{
		userUC:        userUC,
		logger:        logger,
		validate:      validate,
		KafkaProducer: KafkaProducer,
	}
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
	// Validate the user
	if err := u.validate.StructCtx(ctx, user); err != nil {
		u.logger.Errorf("validate.StructCtx", err)
	}
	createdUser, err := u.userUC.Register(ctx, user)

	if err != nil {
		u.logger.Errorf("userUC.CreateUser : %v", err)
		return nil, grpc_error.ErrorResponse(err, "userUC.CreateUser")
	}

	userJSON, err := json.Marshal(createdUser)
	if err != nil {
		u.logger.Errorf("json.Marshal: %v", err)
		// Handle error
	}

	if err := u.KafkaProducer.PublishCreate(ctx, &sarama.ProducerMessage{
		Topic: kafka.CreateUserTopic,
		Value: sarama.StringEncoder(userJSON),
	}); err != nil {
		u.logger.Errorf("KafkaProducer.PublishCreate: %v", err)
	}

	return &userService.CreateUserResponse{User: createdUser.ToProto()}, nil
}
