package main

import (
	"Go-grpc/config"
	"Go-grpc/internal/server"
	"Go-grpc/pkg/jaeger"
	"Go-grpc/pkg/logger"
	"Go-grpc/pkg/postgres"
	"Go-grpc/pkg/redis"
	"log"
	"os"

	"github.com/opentracing/opentracing-go"
)

func main() {
	configPath := config.GetConfigPath(os.Getenv("config"))
	cfg, err := config.GetConfig(configPath)

	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()
	appLogger.Info("Starting user server")
	appLogger.Infof(
		"AppVersion : %s, LogLevel: %s, Mode: %s",
		cfg.GRPCServer.AppVersion,
		cfg.Logger.Level,
		cfg.GRPCServer.Mode,
	)

	appLogger.Infof("Success parsed config: %#v", cfg.GRPCServer.AppVersion)

	pgxConn, err := postgres.NewPgxConn(cfg)

	if err != nil {
		appLogger.Fatal("cannot connect to postgres", err)
	}

	defer pgxConn.Close()

	redisClient := redis.NewRedisClient(cfg)
	appLogger.Info("Redis Connected.")

	tracer, closer, err := jaeger.IntJaeger(cfg)

	if err != nil {
		appLogger.Fatal("cannot create tracer", err)
	}

	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)

	defer closer.Close()

	appLogger.Info("Opentracing connected")

	appLogger.Infof("%-v", pgxConn.Stat())
	appLogger.Infof("%-v", redisClient.PoolStats())

	s := server.NewServer(appLogger, cfg, redisClient, pgxConn, tracer)

	appLogger.Fatal(s.Run())
}
