package middleware

import (
	"strings"

	"nekozanedex/internal/config"
	"nekozanedex/internal/utils"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware - Xác thực JWT token
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		var tokenString string
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}else {
			tokenString, _ = c.Cookie("access_token")
		}

		if tokenString == "" {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		// Validate token
		claims, err := utils.VerifyAccessToken(tokenString, cfg.Jwt.AccessSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		//nếu hợp lệ
		//Set user info vào context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next() // Chạy middleware tiếp theo
	}
}

// Admin Middleware - Chỉ cho phép admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Forbidden(c, "Cần Quyền Admin")
			c.Abort()
			return
		}
		c.Next() // Chạy middleware tiếp theo
	}
}

// RoleMiddleware - Cho phép nhiều roles (linh hoạt hơn AdminMiddleware)
// Usage: RoleMiddleware("admin") hoặc RoleMiddleware("reader", "admin")
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Forbidden(c, "Không tìm thấy role trong token")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Forbidden(c, "Role không hợp lệ")
			c.Abort()
			return
		}

		// Kiểm tra role có trong danh sách cho phép không
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Bạn không có quyền truy cập tài nguyên này")
		c.Abort()
	}
}

// Optional Auth Middleware - Auth không bắt buộc (cho guest)
func OptionalAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			if claims, err := utils.VerifyAccessToken(parts[1], cfg.Jwt.AccessSecret); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("role", claims.Role)
			}
		}

		c.Next() // Chạy middleware tiếp theo
	}
}
