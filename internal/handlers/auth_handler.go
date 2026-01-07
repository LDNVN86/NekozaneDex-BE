package handlers

import (
	"nekozanedex/internal/config"
	"nekozanedex/internal/middleware"
	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService services.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// RegisterRequest - Request body cho đăng ký
type RegisterRequest struct {
	Email    string 	`json:"email" binding:"required,email"`
	Username string 	`json:"username" binding:"required,min=3,max=50"`
	Password string 	`json:"password" binding:"required,min=8"`
}

// LoginRequest - Request body cho đăng nhập
type LoginRequest struct {
	Email    string 	`json:"email" binding:"required,email"`
	Password string 	`json:"password" binding:"required"`
}

// ChangePasswordRequest - Request body cho đổi mật khẩu
type ChangePasswordRequest struct {
	OldPassword string 	`json:"old_password" binding:"required"`
	NewPassword string 	`json:"new_password" binding:"required,min=8"`
}

// RefreshRequest - Request body cho refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register godoc
// @Summary Đăng ký tài khoản mới
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Register Info"
// @Success 201 {object} response.Response
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	user, err := h.authService.Register(req.Email, req.Username, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}

// Login godoc
// @Summary Đăng nhập
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login Info"
// @Success 200 {object} response.Response
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	// Lấy thông tin device - ip user
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokenPair, user, err := h.authService.Login(req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Set tokens vào HttpOnly cookies
	h.setAccessTokenCookie(c, tokenPair.AccessToken)
	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)
	middleware.SetCSRFCookie(c, user.ID.String(), h.getCSRFConfig())

	response.Oke(c, gin.H{
		"access_token": tokenPair.AccessToken,
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// RefreshToken godoc
// @Summary Làm mới access token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Lấy refresh token từ cookie hoặc body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// Thử lấy từ body
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Unauthorized(c, "Refresh token không tồn tại")
			return
		}
		refreshToken = req.RefreshToken
	}

	// Lấy thông tin device
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokenPair, err := h.authService.RefreshToken(refreshToken, userAgent, ipAddress)
	if err != nil {
		// Clear cookie nếu token không hợp lệ
		h.clearRefreshTokenCookie(c)
		response.Unauthorized(c, err.Error())
		return
	}

	// Update tokens cookies
	h.setAccessTokenCookie(c, tokenPair.AccessToken)
	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)

	response.Oke(c, gin.H{
		"access_token": tokenPair.AccessToken,
	})
}

// GetProfile godoc
// @Summary Lấy thông tin profile
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
	})
}

// ChangePassword godoc
// @Summary Đổi mật khẩu
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body ChangePasswordRequest true "Password Info"
// @Success 200 {object} response.Response
// @Router /api/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	err := h.authService.ChangePassword(userID.(uuid.UUID), req.OldPassword, req.NewPassword)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Clear cookie sau khi đổi password (buộc đăng nhập lại)
	h.clearRefreshTokenCookie(c)
	h.clearAccessTokenCookie(c)

	response.Oke(c, gin.H{"message": "Đổi mật khẩu thành công, vui lòng đăng nhập lại"})
}

// Logout godoc
// @Summary Đăng xuất
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Lấy refresh token từ cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		// Revoke token trong DB
		_ = h.authService.Logout(refreshToken)
	}

	// Xóa cookies
	h.clearAccessTokenCookie(c)
	h.clearRefreshTokenCookie(c)
	middleware.ClearCSRFCookie(c, h.getCSRFConfig())

	response.Oke(c, gin.H{"message": "Đăng xuất thành công"})
}

// LogoutAll godoc
// @Summary Đăng xuất tất cả thiết bị
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	if err := h.authService.LogoutAll(userID.(uuid.UUID)); err != nil {
		response.InternalServerError(c, "Không thể đăng xuất")
		return
	}

	// Xóa cookies hiện tại
	h.clearAccessTokenCookie(c)
	h.clearRefreshTokenCookie(c)

	response.Oke(c, gin.H{"message": "Đã đăng xuất khỏi tất cả thiết bị"})
}

// GetSessions godoc
// @Summary Lấy danh sách sessions đang active
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	sessions, err := h.authService.GetActiveSessions(userID.(uuid.UUID))
	if err != nil {
		response.InternalServerError(c, "Không thể lấy danh sách sessions")
		return
	}

	// Map to safe response (không expose token hash)
	var result []gin.H
	for _, s := range sessions {
		result = append(result, gin.H{
			"id":         s.ID,
			"user_agent": s.UserAgent,
			"ip_address": s.IPAddress,
			"created_at": s.CreatedAt,
			"expires_at": s.ExpiresAt,
		})
	}

	response.Oke(c, result)
}

// Helper: Set access token cookie
func (h *AuthHandler) setAccessTokenCookie(c *gin.Context, token string) {
	c.SetCookie(
		"access_token",
		token,
		h.cfg.Jwt.AccessExpireMinutes*60, // Giây
		h.cfg.Cookie.Path,
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		h.cfg.Cookie.HttpOnly,
	)
}

// Helper: Set refresh token cookie
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	c.SetCookie(
		"refresh_token",
		token,
		h.cfg.Cookie.MaxAge,
		h.cfg.Cookie.Path,
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		h.cfg.Cookie.HttpOnly,
	)
}

// Helper: Clear access token cookie
func (h *AuthHandler) clearAccessTokenCookie(c *gin.Context) {
	c.SetCookie(
		"access_token",
		"",
		-1,
		h.cfg.Cookie.Path,
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		h.cfg.Cookie.HttpOnly,
	)
}

// Helper: Clear refresh token cookie
func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		h.cfg.Cookie.Path,
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		h.cfg.Cookie.HttpOnly,
	)
}

// Helper: Get CSRF config from app config
func (h *AuthHandler) getCSRFConfig() middleware.CSRFConfig {
	cfg := middleware.DefaultCSRFConfig()
	cfg.SecretKey = h.cfg.CSRF.SecretKey
	cfg.Secure = h.cfg.App.IsProduction
	cfg.CookieDomain = h.cfg.Cookie.Domain
	return cfg
}
