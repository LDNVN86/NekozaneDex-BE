package handlers

import (
	"fmt"
	"nekozanedex/internal/config"
	"nekozanedex/internal/middleware"
	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService   services.AuthService
	uploadService services.UploadService
	cfg           *config.Config
}

func NewAuthHandler(authService services.AuthService, uploadService services.UploadService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		uploadService: uploadService,
		cfg:           cfg,
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

// UpdateProfileRequest - Request body cho cập nhật profile
type UpdateProfileRequest struct {
	Username     *string `json:"username" binding:"omitempty,min=3,max=50"`
	AvatarURL    *string `json:"avatar_url" binding:"omitempty,url"`
	OldAvatarURL *string `json:"old_avatar_url" binding:"omitempty"`
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

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokenPair, user, err := h.authService.Login(req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

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
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Unauthorized(c, "Refresh token không tồn tại")
			return
		}
		refreshToken = req.RefreshToken
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokenPair, err := h.authService.RefreshToken(refreshToken, userAgent, ipAddress)
	if err != nil {
		h.clearRefreshTokenCookie(c)
		response.Unauthorized(c, err.Error())
		return
	}

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

// UpdateProfile godoc
// @Summary Cập nhật thông tin profile
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body UpdateProfileRequest true "Profile Info"
// @Success 200 {object} response.Response
// @Router /api/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	// At least one field must be provided
	if req.Username == nil && req.AvatarURL == nil {
		response.BadRequest(c, "Không có thông tin để cập nhật")
		return
	}

	user, err := h.authService.UpdateProfile(userID.(uuid.UUID), req.Username, req.AvatarURL)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Delete old avatar from Cloudinary if new avatar was provided
	if req.AvatarURL != nil && req.OldAvatarURL != nil && *req.OldAvatarURL != "" {
		publicID := extractCloudinaryPublicID(*req.OldAvatarURL)
		if publicID != "" {
			if err := h.uploadService.DeleteImage(publicID); err != nil {
				// Log error but don't fail the request
				fmt.Printf("[UpdateProfile] Failed to delete old avatar: %v\n", err)
			} else {
				fmt.Printf("[UpdateProfile] Deleted old avatar: %s\n", publicID)
			}
		}
	}

	response.Oke(c, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
		"created_at": user.CreatedAt,
		"message":    "Cập nhật thông tin thành công",
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

	h.clearRefreshTokenCookie(c)
	h.clearAccessTokenCookie(c)

	response.Oke(c, gin.H{"message": "Đổi mật khẩu thành công, vui lòng đăng nhập lại bạn Nhé"})
}

// Logout godoc
// @Summary Đăng xuất
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		_ = h.authService.Logout(refreshToken)
	}

	h.clearAccessTokenCookie(c)
	h.clearRefreshTokenCookie(c)
	middleware.ClearCSRFCookie(c, h.getCSRFConfig())

	response.Oke(c, gin.H{"message": "Đăng `Xuất` thành công"})
}

// LogoutAll godoc
// @Summary Đăng xuất tất cả thiết bị
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/logout-all [post]
// Này để trưng cho đẹp thôi chứ chả có ma nào dùng🐧🐧
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập bạn Nhé")
		return
	}

	if err := h.authService.LogoutAll(userID.(uuid.UUID)); err != nil {
		response.InternalServerError(c, "Không thể đăng xuất bạn Nhé")
		return
	}

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

func (h *AuthHandler) setAccessTokenCookie(c *gin.Context, token string) {
	c.SetCookie(
		"access_token",
		token,
		h.cfg.Jwt.AccessExpireSeconds,
		h.cfg.Cookie.Path,
		h.cfg.Cookie.Domain,
		h.cfg.Cookie.Secure,
		h.cfg.Cookie.HttpOnly,
	)
}

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

func (h *AuthHandler) getCSRFConfig() middleware.CSRFConfig {
	cfg := middleware.DefaultCSRFConfig()
	cfg.SecretKey = h.cfg.CSRF.SecretKey
	cfg.Secure = h.cfg.App.IsProduction
	cfg.CookieDomain = h.cfg.Cookie.Domain
	return cfg
}

// extractCloudinaryPublicID extracts the public ID from a Cloudinary URL
// Example: https://res.cloudinary.com/xxx/image/upload/v1234/avatars/abc123.webp
// Returns: avatars/abc123
func extractCloudinaryPublicID(url string) string {
	if url == "" {
		return ""
	}

	// Match pattern: /upload/v<version>/<folder>/<filename>.<ext>
	re := regexp.MustCompile(`/upload/v\d+/(.+)\.\w+$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}

	// Alternative pattern without version: /upload/<folder>/<filename>.<ext>
	re2 := regexp.MustCompile(`/upload/(.+)\.\w+$`)
	matches2 := re2.FindStringSubmatch(url)
	if len(matches2) > 1 {
		return matches2[1]
	}

	return ""
}

