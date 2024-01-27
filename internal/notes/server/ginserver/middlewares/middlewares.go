package middlewares

import (
	"time"

	"notes/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware(logg logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqURI := ctx.Request.RequestURI
		reqMethod := ctx.Request.Method
		clientIP := ctx.ClientIP()
		userAgent := ctx.Request.UserAgent()

		startTime := time.Now()

		ctx.Next()

		latency := time.Since(startTime)
		latencyTime := latency.String()
		statusCode := ctx.Writer.Status()

		logg.Info("REST API request",
			zap.String("METHOD", reqMethod),
			zap.String("URI", reqURI),
			zap.Int("STATUS", statusCode),
			zap.String("LATENCY", latencyTime),
			zap.String("Client IP", clientIP),
			zap.String("User Agent", userAgent),
		)
		ctx.Next()
	}
}
