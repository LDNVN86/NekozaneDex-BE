package handlers

import (
	"nekozanedex/internal/config"
	"nekozanedex/internal/middleware"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CSRFHandler struct {
	cfg *config.Config
}

func NewCSRFHandler(cfg *config.Config) *CSRFHandler {
	return &CSRFHandler{cfg: cfg}
}

// GetCSRFToken cấp lại CSRF token cho user đã authenticated
// @Summary Get CSRF Token
// @Description Cấp CSRF token mới cho user đã đăng nhập
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "Token set in cookie"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /auth/csrf-token [get]
func (h *CSRFHandler) GetCSRFToken(c *gin.Context) {
	// Lấy user_id từ context (đã được set bởi AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userIDStr := userID.(uuid.UUID).String()

	// Tạo CSRF config
	csrfCfg := middleware.DefaultCSRFConfig()
	csrfCfg.SecretKey = h.cfg.CSRF.SecretKey
	csrfCfg.Secure = h.cfg.App.IsProduction

	// Set CSRF cookie
	middleware.SetCSRFCookie(c, userIDStr, csrfCfg)

	response.Oke(c, gin.H{
		"message": "CSRF token refreshed successfully",
	})
}
