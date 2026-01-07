package middleware

import (
	"time"

	"nekozanedex/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware - Cấu hình CORS dựa trên environment
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
	}

	// Thêm production origins từ config
	if cfg.App.IsProduction {
		// Production: chỉ cho phép domain chính thức
		allowedOrigins = []string{
			"https://nekozanedex.com",
			"https://www.nekozanedex.com",
		}
	}

	// Staging environment
	if cfg.App.Env == "staging" {
		allowedOrigins = append(allowedOrigins,
			"https://staging.nekozanedex.com",
		)
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// StrictCORSMiddleware - CORS nghiêm ngặt hơn cho sensitive endpoints
func StrictCORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	var allowedOrigins []string

	if cfg.App.IsProduction {
		allowedOrigins = []string{"https://nekozanedex.com"}
	} else {
		allowedOrigins = []string{"http://localhost:3000"}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"POST"}, // Chỉ POST cho auth endpoints
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
	})
}
