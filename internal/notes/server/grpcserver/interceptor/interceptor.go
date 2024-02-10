package interceptor

import (
	"context"
	"notes/internal/pkg/logger"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type UnaryServerInterceptor func(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error)

func LoggingInterceptor(logg logger.Logger) UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logg.Error("cannot get metadata from context")
			return handler(ctx, req)
		}
		userAgent := getUsegAgents(md)
		clientIP := getIP(ctx)

		start := time.Now()

		method := info.FullMethod

		resp, err = handler(ctx, req)

		latency := time.Since(start)

		var statusCode string
		var message string
		if stat, ok := status.FromError(err); ok {
			statusCode = stat.Code().String()
			message = stat.Message()
		}

		logg.Info("GRPC API request",
			zap.String("METHOD", method),
			zap.String("STATUS", statusCode),
			zap.String("Message", message),
			zap.String("LATENCY", latency.String()),
			zap.String("Client IP", clientIP),
			zap.String("User Agent", userAgent),
		)
		return resp, err
	}
}

func getUsegAgents(md metadata.MD) string {
	ua := md.Get("user-agent")

	userAgent := ""
	for _, a := range ua {
		userAgent += a
	}

	return userAgent
}

func getIP(ctx context.Context) string {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	addr := pr.Addr.String()
	return addr
}
