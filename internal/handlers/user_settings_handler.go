package handlers

import (
	"nekozanedex/internal/models"
	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserSettingsHandler struct {
	settingsService services.UserSettingsService
}

func NewUserSettingsHandler(settingsService services.UserSettingsService) *UserSettingsHandler {
	return &UserSettingsHandler{settingsService: settingsService}
}

type UpdateSettingsRequest struct {
	Theme           string  `json:"theme"`
	FontSize        int     `json:"font_size"`
	FontFamily      string  `json:"font_family"`
	LineHeight      float64 `json:"line_height"`
	ReadingBg       string  `json:"reading_bg"`
	AutoScrollSpeed int     `json:"auto_scroll_speed"`
}

// GetMySettings godoc
// @Summary Lấy settings của tôi
// @Tags Settings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/settings [get]
func (h *UserSettingsHandler) GetMySettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	settings, err := h.settingsService.GetSettings(userID.(uuid.UUID))
	if err != nil {
		response.InternalServerError(c, "Không thể lấy settings")
		return
	}

	response.Oke(c, settings)
}

// UpdateMySettings godoc
// @Summary Cập nhật settings của tôi
// @Tags Settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body UpdateSettingsRequest true "Settings Info"
// @Success 200 {object} response.Response
// @Router /api/settings [put]
func (h *UserSettingsHandler) UpdateMySettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	if req.Theme != "" && req.Theme != "light" && req.Theme != "dark" && req.Theme != "system" {
		response.BadRequest(c, "Theme không hợp lệ (light/dark/system)")
		return
	}

	if req.FontSize != 0 && (req.FontSize < 12 || req.FontSize > 32) {
		response.BadRequest(c, "Font size phải từ 12-32")
		return
	}

	validBgs := []string{"white", "sepia", "dark", "black"}
	if req.ReadingBg != "" {
		valid := false
		for _, bg := range validBgs {
			if req.ReadingBg == bg {
				valid = true
				break
			}
		}
		if !valid {
			response.BadRequest(c, "Reading background không hợp lệ")
			return
		}
	}

	updates := &models.UserSettings{
		Theme:           req.Theme,
		FontSize:        req.FontSize,
		FontFamily:      req.FontFamily,
		LineHeight:      req.LineHeight,
		ReadingBg:       req.ReadingBg,
		AutoScrollSpeed: req.AutoScrollSpeed,
	}

	settings, err := h.settingsService.UpdateSettings(userID.(uuid.UUID), updates)
	if err != nil {
		response.InternalServerError(c, "Không thể cập nhật settings")
		return
	}

	response.Oke(c, settings)
}
