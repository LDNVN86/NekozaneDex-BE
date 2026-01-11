package middleware

import (
	"time"

	"nekozanedex/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware - Cấu hình CORS dựa trên environment
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	var allowedOrigins []string

	switch {
	case cfg.App.IsProduction:
		allowedOrigins = cfg.CORS.ProdOrigins
	case cfg.App.Env == "staging":
		allowedOrigins = append(cfg.CORS.DevOrigins, cfg.CORS.StagingOrigins...)
	default:
		allowedOrigins = cfg.CORS.DevOrigins
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
		// Production: chỉ cho phép origin đầu tiên (main domain)
		if len(cfg.CORS.ProdOrigins) > 0 {
			allowedOrigins = []string{cfg.CORS.ProdOrigins[0]}
		}
	} else {
		// Dev: chỉ cho phép origin đầu tiên
		if len(cfg.CORS.DevOrigins) > 0 {
			allowedOrigins = []string{cfg.CORS.DevOrigins[0]}
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"POST"}, // Chỉ POST cho auth endpoints
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
	})
}
