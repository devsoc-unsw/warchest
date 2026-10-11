package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Req_logger() gin.HandlerFunc {
	return func (c *gin.Context) {
		start_time := time.Now()
		
		// Pre-handler phase add code here
		c.Next()
		// Post-handler phase

		latency := time.Since(start_time)
		slog.Info("latency", latency)
	}
}


// TODO
// Log into slog details of request
	// Break down the details from c 
	// Society_id, request, latency