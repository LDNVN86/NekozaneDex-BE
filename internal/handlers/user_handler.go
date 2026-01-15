package handlers

import (
	"nekozanedex/internal/repositories"
	"nekozanedex/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo repositories.UserRepository
}

func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin reader"`
}
type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// SearchUsers godoc
// @Summary Search users by username (for @mention autocomplete)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param q query string true "Username query"
// @Param limit query int false "Max results" default(5)
// @Success 200 {object} response.Response
// @Router /api/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	if limit < 1 || limit > 10 {
		limit = 5
	}

	users, err := h.userRepo.SearchUsersByUsername(query, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể tìm kiếm người dùng")
		return
	}

	response.Oke(c, users)
}

// GetAllUsersAdmin godoc
// @Summary Get all users with pagination and search
// @Tags Admin Users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search keyword"
// @Success 200 {object} response.Response
// @Router /api/admin/users [get]
func (h *UserHandler) GetAllUsersAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.userRepo.SearchUsersAdmin(search, page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy danh sách người dùng")
		return
	}

	response.PaginatedResponse(c, users, page, limit, total)
}

// UpdateUserRole godoc
// @Summary Update user role
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UpdateRoleRequest true "Role info"
// @Success 200 {object} response.Response
// @Router /api/admin/users/{id}/role [put]
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID người dùng không hợp lệ")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	user, err := h.userRepo.FindUserByID(userID)
	if err != nil {
		response.NotFound(c, "Không tìm thấy người dùng")
		return
	}

	if user.Role == "admin" && req.Role == "reader" {
		response.Forbidden(c, "Không thể hạ quyền Admin xuống Reader")
		return
	}

	user.Role = req.Role
	if err := h.userRepo.UpdateUser(user); err != nil {
		response.InternalServerError(c, "Không thể cập nhật role")
		return
	}

	response.Oke(c, gin.H{
		"id":      user.ID,
		"role":    user.Role,
		"message": "Cập nhật role thành công",
	})
}

// ToggleUserStatus godoc
// @Summary Toggle user active status
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UpdateStatusRequest true "Status info"
// @Success 200 {object} response.Response
// @Router /api/admin/users/{id}/status [put]
func (h *UserHandler) ToggleUserStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID người dùng không hợp lệ")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	user, err := h.userRepo.FindUserByID(userID)
	if err != nil {
		response.NotFound(c, "Không tìm thấy người dùng")
		return
	}

	if user.Role == "admin" && !req.IsActive {
		response.Forbidden(c, "Không thể vô hiệu hóa tài khoản Admin")
		return
	}

	user.IsActive = req.IsActive
	if err := h.userRepo.UpdateUser(user); err != nil {
		response.InternalServerError(c, "Không thể cập nhật trạng thái")
		return
	}

	response.Oke(c, gin.H{
		"id":        user.ID,
		"is_active": user.IsActive,
		"message":   "Cập nhật trạng thái thành công",
	})
}

type AdminUpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role" binding:"omitempty,oneof=admin reader"`
}

type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
// @Summary Update user info (username, email, role)
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body AdminUpdateUserRequest true "User info"
// @Success 200 {object} response.Response
// @Router /api/admin/users/{id} [put]
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID người dùng không hợp lệ")
		return
	}

	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	user, err := h.userRepo.FindUserByID(userID)
	if err != nil {
		response.NotFound(c, "Không tìm thấy người dùng")
		return
	}

	if user.Role == "admin" && req.Role == "reader" {
		response.Forbidden(c, "Không thể hạ quyền Admin xuống Reader")
		return
	}

	if req.Username != "" {
		existing, _ := h.userRepo.FindUserByUsername(req.Username)
		if existing != nil && existing.ID != user.ID {
			response.Conflict(c, "Username đã tồn tại")
			return
		}
		user.Username = req.Username
	}

	if req.Email != "" {
		existing, _ := h.userRepo.FindUserByEmail(req.Email)
		if existing != nil && existing.ID != user.ID {
			response.Conflict(c, "Email đã tồn tại")
			return
		}
		user.Email = req.Email
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if err := h.userRepo.UpdateUser(user); err != nil {
		response.InternalServerError(c, "Không thể cập nhật người dùng")
		return
	}

	response.Oke(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"message":  "Cập nhật thông tin thành công",
	})
}

// AdminResetPassword godoc
// @Summary Reset user password
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body AdminResetPasswordRequest true "New password"
// @Success 200 {object} response.Response
// @Router /api/admin/users/{id}/password [put]
func (h *UserHandler) AdminResetPassword(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID người dùng không hợp lệ")
		return
	}

	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Mật khẩu phải có ít nhất 8 ký tự")
		return
	}

	user, err := h.userRepo.FindUserByID(userID)
	if err != nil {
		response.NotFound(c, "Không tìm thấy người dùng")
		return
	}

	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		response.InternalServerError(c, "Không thể mã hóa mật khẩu")
		return
	}

	user.PasswordHash = hashedPassword
	if err := h.userRepo.UpdateUser(user); err != nil {
		response.InternalServerError(c, "Không thể cập nhật mật khẩu")
		return
	}

	response.Oke(c, gin.H{
		"id":      user.ID,
		"message": "Đã đổi mật khẩu thành công",
	})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
