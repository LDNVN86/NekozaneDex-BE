package middleware

import (
	"strings"

	"nekozanedex/internal/config"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders - Thêm các security headers để bảo vệ ứng dụng
func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		frameAncestors := strings.TrimSpace(cfg.Security.FrameAncestors)
		if frameAncestors == "" {
			frameAncestors = "'self'"
		}
		switch frameAncestors {
		case "self":
			frameAncestors = "'self'"
		case "none":
			frameAncestors = "'none'"
		}

		xFrameOptions := ""
		switch frameAncestors {
		case "'none'":
			xFrameOptions = "DENY"
		case "'self'":
			xFrameOptions = "SAMEORIGIN"
		}

		// Prevent clickjacking - allow config-driven frame ancestors
		if xFrameOptions != "" {
			c.Header("X-Frame-Options", xFrameOptions)
		}

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS Protection (legacy nhưng vẫn hữu ích cho IE)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer Policy - không gửi referrer khi chuyển sang HTTP
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy - vô hiệu hóa các features không cần thiết
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy - hạn chế nguồn resources
		// Note: Cần điều chỉnh theo nhu cầu thực tế của frontend
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+ // Cho phép inline scripts (Next.js cần)
				"style-src 'self' 'unsafe-inline'; "+ // Cho phép inline styles
				"img-src 'self' data: https: blob:; "+ // Cho phép images từ HTTPS
				"font-src 'self' data:; "+
				"connect-src 'self' ws: wss: https:; "+ // Cho phép WebSocket và HTTPS
				"frame-ancestors "+frameAncestors)

		c.Next()
	}
}

// ProductionSecurityHeaders - Headers bổ sung cho production
func ProductionSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HSTS - buộc sử dụng HTTPS (chỉ cho production)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		c.Next()
	}
}
