package middleware

import (
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
