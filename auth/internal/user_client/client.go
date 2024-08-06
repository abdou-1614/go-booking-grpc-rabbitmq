package userclient

import (
	"auth/config"
	"auth/internal/interceptors"
	"context"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	traceutils "github.com/opentracing-contrib/go-grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	backoffLinear = 100 * time.Millisecond
)

func NewUserServiceConn(ctx context.Context, cf *config.Config, interceptor *interceptors.InterceptorManager) (*grpc.ClientConn, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(backoffLinear)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
	}

	userGRPCConn, err := grpc.DialContext(
		ctx,
		cf.GRPCServer.UserGrpcServicePort,
		grpc.WithUnaryInterceptor(traceutils.OpenTracingClientInterceptor(interceptor.GetTracer())),
		grpc.WithUnaryInterceptor(interceptor.GetInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)),
	)
	if err != nil {
		return nil, err
	}
	return userGRPCConn, nil
}
