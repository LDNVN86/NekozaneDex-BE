package services

import (
	"errors"
	"time"

	"nekozanedex/internal/config"
	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"
	"nekozanedex/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(email, username, password string) (*models.User, error)
	Login(email, password, userAgent, ipAddress string) (*utils.TokenPair, *models.User, error)
	RefreshToken(refreshToken, userAgent, ipAddress string) (*utils.TokenPair, error)
	Logout(refreshToken string) error
	LogoutAll(userID uuid.UUID) error
	GetUserByID(id uuid.UUID) (*models.User, error)
	UpdateProfile(userID uuid.UUID, username, avatarURL *string) (*models.User, error)
	ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error
	GetActiveSessions(userID uuid.UUID) ([]models.RefreshToken, error)
}

type authService struct {
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	cfg              *config.Config
}


//constructor
func NewAuthService(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
	}
}

// Register - Đăng ký tài khoản mới
func (s *authService) Register(email, username, password string) (*models.User, error) {
	// Sanitize inputs
	email = utils.SanitizeInput(email)
	username = utils.SanitizeInput(username)

	// Validate email
	if !utils.ValidateEmail(email) {
		return nil, errors.New("email không hợp lệ")
	}

	// Validate username
	if err := utils.ValidateUsername(username); err != nil {
		return nil, err
	}

	// Kiểm tra email đã tồn tại chưa
	existingUser, _ := s.userRepo.FindUserByEmail(email)
	if existingUser != nil {
		return nil, errors.New("email đã được sử dụng")
	}

	// Kiểm tra username đã tồn tại chưa
	existingUser, _ = s.userRepo.FindUserByUsername(username)
	if existingUser != nil {
		return nil, errors.New("username đã được sử dụng")
	}

	// Validate password strength (sử dụng policy mới)
	if err := utils.ValidatePasswordDefault(password); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("không thể mã hóa mật khẩu")
	}

	// Tạo user mới
	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         "reader",
		IsActive:     true,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, errors.New("không thể tạo tài khoản")
	}

	return user, nil
}

// Login - Đăng nhập với refresh token lưu DB
func (s *authService) Login(email, password, userAgent, ipAddress string) (*utils.TokenPair, *models.User, error) {
	// Tìm user theo email
	user, err := s.userRepo.FindUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("email hoặc mật khẩu không đúng")
		}
		return nil, nil, err
	}

	// Kiểm tra tài khoản có active không
	if !user.IsActive {
		return nil, nil, errors.New("tài khoản đã bị khóa")
	}

	// Verify password
	match, err := utils.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return nil, nil, errors.New("email hoặc mật khẩu không đúng")
	}

	// Generate tokens và lưu refresh token vào DB
	tokenPair, err := s.generateAndStoreTokens(user, userAgent, ipAddress)
	if err != nil {
		return nil, nil, errors.New("không thể tạo token")
	}

	return tokenPair, user, nil
}

// RefreshToken - Làm mới token với rotation (tạo mới cả refresh token)
// Implements Refresh Token Rotation with Reuse Detection
func (s *authService) RefreshToken(refreshToken, userAgent, ipAddress string) (*utils.TokenPair, error) {
	// Hash token để tìm trong DB
	tokenHash := utils.HashToken(refreshToken)

	// Tìm token trong DB
	storedToken, err := s.refreshTokenRepo.FindByHash(tokenHash)
	if err != nil {
		return nil, errors.New("refresh token không hợp lệ")
	}

	// REUSE DETECTION: Nếu token đã bị revoke trước đó
	// -> Có thể bị đánh cắp, revoke TẤT CẢ tokens của user này
	if storedToken.IsRevoked() {
		// Security: Revoke all tokens for this user
		_ = s.refreshTokenRepo.RevokeAllByUser(storedToken.UserID)
		return nil, errors.New("token đã bị thu hồi - vui lòng đăng nhập lại")
	}

	// Kiểm tra token còn hạn không
	if storedToken.IsExpired() {
		return nil, errors.New("refresh token đã hết hạn")
	}

	// Lấy thông tin user
	user, err := s.userRepo.FindUserByID(storedToken.UserID)
	if err != nil {
		return nil, errors.New("user không tồn tại")
	}

	if !user.IsActive {
		return nil, errors.New("tài khoản đã bị khóa")
	}

	// ROTATION: Revoke token cũ
	if err := s.refreshTokenRepo.Revoke(storedToken.ID); err != nil {
		return nil, errors.New("không thể thu hồi token cũ")
	}

	// Tạo token mới
	tokenPair, err := s.generateAndStoreTokens(user, userAgent, ipAddress)
	if err != nil {
		return nil, errors.New("không thể tạo token mới")
	}

	return tokenPair, nil
}

// Logout - Đăng xuất (revoke refresh token hiện tại)
func (s *authService) Logout(refreshToken string) error {
	tokenHash := utils.HashToken(refreshToken)
	return s.refreshTokenRepo.RevokeByHash(tokenHash)
}

// LogoutAll - Đăng xuất tất cả thiết bị
func (s *authService) LogoutAll(userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeAllByUser(userID)
}

// GetUserByID - Lấy thông tin user theo ID
func (s *authService) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindUserByID(id)
	if err != nil {
		return nil, errors.New("user không tồn tại")
	}
	return user, nil
}

// UpdateProfile - Cập nhật thông tin profile
func (s *authService) UpdateProfile(userID uuid.UUID, username, avatarURL *string) (*models.User, error) {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user không tồn tại")
	}

	// Update username if provided
	if username != nil {
		sanitizedUsername := utils.SanitizeInput(*username)
		if err := utils.ValidateUsername(sanitizedUsername); err != nil {
			return nil, err
		}

		// Check if username is already taken by another user
		existingUser, _ := s.userRepo.FindUserByUsername(sanitizedUsername)
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("username đã được sử dụng")
		}

		user.Username = sanitizedUsername
	}

	// Update avatar URL if provided
	if avatarURL != nil {
		user.AvatarURL = avatarURL
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, errors.New("không thể cập nhật thông tin")
	}

	return user, nil
}

// ChangePassword - Đổi mật khẩu và revoke tất cả tokens
func (s *authService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return errors.New("user không tồn tại")
	}

	// Verify old password
	match, err := utils.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil || !match {
		return errors.New("mật khẩu cũ không đúng")
	}

	// Validate new password (sử dụng policy mới)
	if err := utils.ValidatePasswordDefault(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("không thể mã hóa mật khẩu")
	}

	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		return err
	}

	// Revoke tất cả refresh tokens (force re-login)
	return s.refreshTokenRepo.RevokeAllByUser(userID)
}

// GetActiveSessions - Lấy danh sách sessions đang active
func (s *authService) GetActiveSessions(userID uuid.UUID) ([]models.RefreshToken, error) {
	return s.refreshTokenRepo.GetActiveByUser(userID)
}

// Helper: Generate tokens và lưu refresh token vào DB
func (s *authService) generateAndStoreTokens(user *models.User, userAgent, ipAddress string) (*utils.TokenPair, error) {
	// Generate access token
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Role,
		s.cfg.Jwt.AccessSecret,
		s.cfg.Jwt.AccessExpireSeconds, // In seconds
	)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		s.cfg.Jwt.RefreshSecret,
		s.cfg.Jwt.RefreshExpireDays,
	)
	if err != nil {
		return nil, err
	}

	// Hash và lưu refresh token vào DB
	tokenHash := utils.HashToken(refreshToken)
	expiresAt := time.Now().Add(time.Duration(s.cfg.Jwt.RefreshExpireDays) * 24 * time.Hour)

	storedToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		UserAgent: &userAgent,
		IPAddress: &ipAddress,
	}

	if err := s.refreshTokenRepo.Create(storedToken); err != nil {
		return nil, err
	}

	return &utils.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
