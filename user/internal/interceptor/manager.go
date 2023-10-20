package interceptor

import (
	"Go-grpc/config"
	"Go-grpc/pkg/logger"
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type InterceptorManager struct {
	logger logger.Loggor
	config *config.Config
	tracer opentracing.Tracer
}

func NewInterceptorManger(cfg *config.Config, tracer opentracing.Tracer, logger logger.Loggor) *InterceptorManager {
	return &InterceptorManager{
		logger: logger,
		config: cfg,
		tracer: tracer,
	}
}

func (im *InterceptorManager) Logger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)

	replay, err := handler(ctx, req)

	im.logger.Infof("METHOD: %v, TIME: %v, METADATA: %v, ERR: %v", info.FullMethod, time.Since(start), md, err)

	return replay, err
}

func (im *InterceptorManager) GetInterceptor() func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		im.logger.Infof("call=%v req=%#v reply=%#v time=%v err=%v",
			method, req, reply, time.Since(start), err)
		return err
	}
}

func (im *InterceptorManager) GetTracer() opentracing.Tracer {
	return im.tracer
}
