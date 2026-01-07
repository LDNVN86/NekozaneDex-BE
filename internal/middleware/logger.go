package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware - Log request/response
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // Bắt đầu thời gian
		path := c.Request.URL.Path // Đường dẫn
		method := c.Request.Method // Phương thức

		// Process request - Chạy middleware
		c.Next()

		// Log sau khi response
		latency := time.Since(start) // Thời gian trễ
		status := c.Writer.Status() // Trạng thái
		clientIP := c.ClientIP() // IP client

		log.Printf("[%s] %s %s | %d | %v | %s",
			method,
			path,
			c.Request.URL.RawQuery,
			status,
			latency,
			clientIP,
		)
	}
}
