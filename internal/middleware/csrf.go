package middleware

import (
	"nekozanedex/internal/utils"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CSRFConfig struct {
	CookieName   string
	HeaderName   string
	CookiePath   string
	CookieDomain string
	Secure       bool
	MaxAge       int
	SecretKey    string 
	ExcludePaths []string
}

func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		CookieName:   "csrf_token",
		HeaderName:   "X-CSRF-Token",
		CookiePath:   "/",
		CookieDomain: "",
		Secure:       false,
		MaxAge:       86400, // 24 hours - synced with token expiry
		SecretKey:    "default-dev-secret-change-this", // Override từ config
		ExcludePaths: []string{
			"/api/auth/login",
			"/api/auth/register",
			"/api/auth/refresh",
			"/api/admin/upload",         // File uploads already have auth
			"/api/admin/upload/chapter", // File uploads already have auth
		},
	}
}

func CSRFMiddleware(cfg CSRFConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, path := range cfg.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(cfg.CookieName)
		if err != nil || cookieToken == "" {
			response.Forbidden(c, "CSRF token missing from cookie")
			c.Abort()
			return
		}

		headerToken := c.GetHeader(cfg.HeaderName)
		if headerToken == "" {
			response.Forbidden(c, "CSRF token missing from header")
			c.Abort()
			return
		}

		if cookieToken != headerToken {
			response.Forbidden(c, "CSRF token mismatch")
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if exists {
			userIDStr := userID.(uuid.UUID).String()
			valid, err := utils.ValidateCSRFToken(cookieToken, userIDStr, cfg.SecretKey)
			if !valid {
				response.Forbidden(c, "CSRF validation failed: "+err.Error())
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func SetCSRFCookie(c *gin.Context, userID string, cfg CSRFConfig) {
	token := utils.GenerateCSRFToken(userID, cfg.SecretKey)
	c.SetCookie(
		cfg.CookieName,
		token,
		cfg.MaxAge,
		cfg.CookiePath,
		cfg.CookieDomain,
		cfg.Secure,
		false,
	)
	c.Header(cfg.HeaderName, token)
}

func ClearCSRFCookie(c *gin.Context, cfg CSRFConfig) {
	c.SetCookie(
		cfg.CookieName,
		"",
		-1,
		cfg.CookiePath,
		cfg.CookieDomain,
		cfg.Secure,
		false,
	)
}