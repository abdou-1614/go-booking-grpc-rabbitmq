package server

import (
	"Go-grpc/config"
	"Go-grpc/internal/interceptor"
	userGRPC "Go-grpc/internal/user/delivery/grpc"
	"Go-grpc/internal/user/delivery/rabbitmq"
	"Go-grpc/internal/user/repository"
	"Go-grpc/internal/user/usecase"
	userGRPCService "Go-grpc/pb"
	"Go-grpc/pkg/logger"
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	certFile          = "ssl/server.crt"
	keyFile           = "ssl/server.pem"
	maxHeaderBytes    = 1 << 20
	userCachePrefix   = "users:"
	userCacheDuration = time.Minute * 15
)

type HelloService struct {
	userGRPCService.UnimplementedHelloServiceServer
}

func (s *HelloService) SayHello(ctx context.Context, in *userGRPCService.HelloRequest) (*userGRPCService.HelloResponse, error) {
	return &userGRPCService.HelloResponse{
		Message: "Hello From the Server !",
		Name:    in.Name,
	}, nil
}

type Server struct {
	userGRPCService.UnimplementedUserServiceServer
	logger    logger.Loggor
	cfg       *config.Config
	redisConn *redis.Client
	pgxPool   *pgxpool.Pool
	tracer    opentracing.Tracer
}

func NewServer(logger logger.Loggor, cfg *config.Config, redisConn *redis.Client, pgxPool *pgxpool.Pool, tracer opentracing.Tracer) *Server {
	server := &Server{
		logger:    logger,
		cfg:       cfg,
		redisConn: redisConn,
		pgxPool:   pgxPool,
		tracer:    tracer,
	}

	return server
}

func (s *Server) Run() error {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	im := interceptor.NewInterceptorManger(s.cfg, s.tracer, s.logger)
	validate := validator.New()

	userPublisher, err := rabbitmq.NewUserPublisher(s.cfg, s.logger)

	if err != nil {
		return errors.Wrap(err, "rabbitmq.NewUserPubliser")
	}

	userPGRepository := repository.NewUserPGRepository(s.pgxPool)
	userRedisRepository := repository.NewUserRedisRepository(s.redisConn, userCachePrefix, userCacheDuration)
	userUseCase := usecase.NewUserUseCase(userPGRepository, s.logger, userRedisRepository, userPublisher)

	userConsumer := rabbitmq.NewUserConsumer(s.cfg, s.logger, userUseCase)

	if err := userConsumer.Dial(); err != nil {
		return errors.Wrap(err, "userConsumer.Dial")
	}

	avatarChan, err := userConsumer.CreateExchangeAndQueue(rabbitmq.UserExchange, rabbitmq.AvatarQueueName, rabbitmq.AvatarsBindingKey)

	if err != nil {
		return errors.Wrap(err, "userConsumer.CreateExchangeAndQueue")
	}

	defer avatarChan.Close()

	userConsumer.RunConsumers(ctx, cancel)

	l, err := net.Listen("tcp", s.cfg.GRPCServer.Port)

	if err != nil {
		return err
	}

	defer l.Close()

	server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: s.cfg.GRPCServer.MaxConnectionIdle * time.Minute,
		Timeout:           s.cfg.GRPCServer.Timeout * time.Second,
		MaxConnectionAge:  s.cfg.GRPCServer.MaxConnectionAge * time.Minute,
		Time:              s.cfg.GRPCServer.Timeout * time.Minute,
	}),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpcrecovery.UnaryServerInterceptor(),
			im.Logger,
		),
	)

	userService := userGRPC.NewUserGRPCService(userUseCase, s.logger, validate)
	userGRPCService.RegisterUserServiceServer(server, userService)
	userGRPCService.RegisterHelloServiceServer(server, new(HelloService))
	grpc_prometheus.Register(server)

	go func() {
		s.logger.Infof("GRPC Server is listening on port: %v", s.cfg.GRPCServer.Port)
		s.logger.Fatal(server.Serve(l))
	}()

	if s.cfg.GRPCServer.Mode != "Production" {
		reflection.Register(server)
	}

	quite := make(chan os.Signal, 1)

	signal.Notify(quite, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quite:
		s.logger.Errorf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		s.logger.Errorf("ctx.Done: %v", done)
	}

	s.logger.Info("Server Exited Properly")

	server.GracefulStop()

	s.logger.Info("Server Exited Properly")

	return nil
}

func (s *Server) RunGateway() error {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	validate := validator.New()

	userPublisher, err := rabbitmq.NewUserPublisher(s.cfg, s.logger)

	if err != nil {
		return errors.Wrap(err, "rabbitmq.NewUserPubliser")
	}

	userPGRepository := repository.NewUserPGRepository(s.pgxPool)
	userRedisRepository := repository.NewUserRedisRepository(s.redisConn, userCachePrefix, userCacheDuration)
	userUseCase := usecase.NewUserUseCase(userPGRepository, s.logger, userRedisRepository, userPublisher)

	userConsumer := rabbitmq.NewUserConsumer(s.cfg, s.logger, userUseCase)

	if err := userConsumer.Dial(); err != nil {
		return errors.Wrap(err, "userConsumer.Dial")
	}

	avatarChan, err := userConsumer.CreateExchangeAndQueue(rabbitmq.UserExchange, rabbitmq.AvatarQueueName, rabbitmq.AvatarsBindingKey)

	if err != nil {
		return errors.Wrap(err, "userConsumer.CreateExchangeAndQueue")
	}

	defer avatarChan.Close()

	userConsumer.RunConsumers(ctx, cancel)

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)

	userService := userGRPC.NewUserGRPCService(userUseCase, s.logger, validate)

	err = userGRPCService.RegisterUserServiceHandlerServer(ctx, grpcMux, userService)
	if err != nil {
		s.logger.Errorf("cannot register handler server : %v", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", grpcMux)

	l, err := net.Listen("tcp", s.cfg.HttpServer.Port)

	if err != nil {
		return err
	}

	defer l.Close()

	go func() {
		s.logger.Infof("Http Server is listening on port: %v", s.cfg.HttpServer.Port)
		s.logger.Fatal(http.Serve(l, mux))
	}()

	quite := make(chan os.Signal, 1)

	signal.Notify(quite, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quite:
		s.logger.Errorf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		s.logger.Errorf("ctx.Done: %v", done)
	}

	s.logger.Info("Server Exited Properly")

	return nil
}
