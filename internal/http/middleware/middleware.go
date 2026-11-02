package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID ensures any coming HTTP request has a request ID
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}

		ctx.Writer.Header().Set("X-Request-ID", id)
		ctx.Set("request_id", id)
	}
}

// Logger logs info about http requests
func Logger(log *slog.Logger) gin.HandlerFunc {
	//customize gin's logging format
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.Info("http request",
			"request_id", param.Keys["request_id"], //Request ID
			"method", param.Method, //HTTP method
			"path", param.Path, // Request path
			"status", param.StatusCode, // Response status code
			"latency", param.Latency, // Request Latency
			"client_ip", param.ClientIP, // Client IP address
		)

		return ""
	})
}
