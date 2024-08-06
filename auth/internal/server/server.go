package server

import (
	"Go-grpc/pkg/logger"
	"auth/config"
	"auth/internal/auth/repository"
	"auth/internal/auth/token"
	"auth/internal/auth/usecase"
	"auth/internal/interceptors"
	userclient "auth/internal/user_client"
	SessionGRPCService "auth/pb/sessions"
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authGRPC "auth/internal/auth/delivery"
	grpcClient "auth/internal/user_client"
	authServer "auth/pb/sessions"
	userService "auth/pb/user"

	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

type Server struct {
	SessionGRPCService.UnimplementedAuthServiceServer
	logger  logger.Loggor
	cfg     *config.Config
	pgxPool *pgxpool.Pool
	tracer  opentracing.Tracer
}

func NewServer(logger logger.Loggor, cfg *config.Config, pgxPool *pgxpool.Pool, tracer opentracing.Tracer) *Server {
	server := &Server{
		logger:  logger,
		cfg:     cfg,
		pgxPool: pgxPool,
		tracer:  tracer,
	}

	return server
}

func (s *Server) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	im := interceptors.NewInterceptorManager(s.logger, s.cfg, s.tracer)

	authPGRepository := repository.NewSessionPGReposity(s.pgxPool)

	userConn, err := grpcClient.NewUserServiceConn(ctx, s.cfg, im)
	if err != nil {
		return err
	}

	defer userConn.Close()

	userClient := userService.NewUserServiceClient(userConn)

	tokenMaker, err := token.NewPasetoMaker(s.cfg.GRPCServer.TokenSymmetricKey)

	if err != nil {
		return err
	}

	authUseCase := usecase.NewSessionUseCase(authPGRepository, s.logger, userClient, s.cfg, tokenMaker, s.tracer)

	l, err := net.Listen("tcp", s.cfg.GRPCServer.Port)
	if err != nil {
		return err
	}
	defer l.Close()

	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
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

	authService := authGRPC.NewSessionGRPCService(authUseCase, s.logger)
	authServer.RegisterAuthServiceServer(server, authService)

	grpc_prometheus.Register(server)

	go func() {
		s.logger.Infof("GRPC Server is listening on port: %v", s.cfg.GRPCServer.Port)
		s.logger.Fatal(server.Serve(l))
	}()

	if s.cfg.GRPCServer.Mode != "Production" {
		reflection.Register(server)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
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

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)

	im := interceptors.NewInterceptorManager(s.logger, s.cfg, s.tracer)

	userConn, err := userclient.NewUserServiceConn(ctx, s.cfg, im)

	if err != nil {
		return err
	}

	defer userConn.Close()

	userClient := userService.NewUserServiceClient(userConn)

	authPGRepository := repository.NewSessionPGReposity(s.pgxPool)
	tokenMaker, err := token.NewPasetoMaker(s.cfg.GRPCServer.TokenSymmetricKey)

	if err != nil {
		return err
	}

	authUseCase := usecase.NewSessionUseCase(authPGRepository, s.logger, userClient, s.cfg, tokenMaker, s.tracer)

	authService := authGRPC.NewSessionGRPCService(authUseCase, s.logger)
	err = authServer.RegisterAuthServiceHandlerServer(ctx, grpcMux, authService)

	if err != nil {
		s.logger.Errorf("cannot register handler server: %v", err)
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	httpServer := &http.Server{
		Addr:         s.cfg.HttpServer.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		s.logger.Infof("HTTP Server is listening on port: %v", s.cfg.HttpServer.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("failed to listen and serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		s.logger.Info("Shutting down HTTP server...")
	case done := <-ctx.Done():
		s.logger.Errorf("ctx.Done: %v", done)
	}

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctxShutDown); err != nil {
		s.logger.Fatalf("HTTP server Shutdown failed: %v", err)
	}

	s.logger.Info("HTTP Server Exited Properly")
	return nil

}
