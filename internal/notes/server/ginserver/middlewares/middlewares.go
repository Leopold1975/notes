package middlewares

import (
	"notes/internal/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
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

		logg.Infof("REST API request	METHOD %s	URI %s	STATUS %d	Latency %s	Client IP %s	User Agent %s\n",
			reqMethod,
			reqURI,
			statusCode,
			latencyTime,
			clientIP,
			userAgent,
		)
		ctx.Next()
	}
}
