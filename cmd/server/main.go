package main

import (
	"log"

	"nekozanedex/internal/config"
	"nekozanedex/internal/database"
	"nekozanedex/internal/handlers"
	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"
	"nekozanedex/internal/routes"
	"nekozanedex/internal/services"

	_ "nekozanedex/docs" // Swagger docs

	"github.com/gin-gonic/gin"
)

// @title           Nekozanedex API
// @version         1.0
// @description     API cho nền tảng đọc truyện web novel Nekozanedex
//
// @host      localhost:8080
// @BasePath  /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Format: "Bearer {token}"

func main() {
	// Load configuration - Load cấu hình
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Không thể load config:", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Connect database
	db, err := database.ConnectDB(&cfg.Database)
	if err != nil {
		log.Fatal("Không thể kết nối database:", err)
	}

	// Auto migrate models - Tự động migrate model
	if err := db.AutoMigrate(
		&models.User{},
		&models.Story{},
		&models.Genre{},
		&models.Chapter{},
		&models.BookMark{},
		&models.ReadingHistory{},
		&models.Comment{},
		&models.Notification{},
		&models.ChatMessage{},
		&models.UserSettings{},
		&models.TypoReport{},
		&models.StoryView{},
		&models.RefreshToken{}, 
	); err != nil {
		log.Fatal("Không thể migrate database:", err)
	}
	log.Println("Database Đã Migrate Thành Công")

	// Initialize repositories - Khởi tạo repository
	userRepo := repositories.NewUserRepository(db)
	storyRepo := repositories.NewStoryRepository(db)
	chapterRepo := repositories.NewChapterRepository(db)
	genreRepo := repositories.NewGenreRepository(db)
	bookmarkRepo := repositories.NewBookmarkRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db) // Thêm RefreshToken repo

	// Initialize services - Khởi tạo service
	authService := services.NewAuthService(userRepo, refreshTokenRepo, cfg) // Cập nhật với refreshTokenRepo
	storyService := services.NewStoryService(storyRepo, genreRepo)
	chapterService := services.NewChapterService(chapterRepo, storyRepo)
	bookmarkService := services.NewBookmarkService(bookmarkRepo, storyRepo)
	commentService := services.NewCommentService(commentRepo, storyRepo, chapterRepo)
	notificationService := services.NewNotificationService(notificationRepo)

	// Initialize upload service (optional - requires Cloudinary config)
	var uploadHandler *handlers.UploadHandler
	uploadService, err := services.NewUploadService(cfg)
	if err != nil {
		log.Printf("⚠️ Upload service not initialized: %v", err)
		log.Println("💡 Add CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET to .env")
	} else {
		uploadHandler = handlers.NewUploadHandler(uploadService)
		log.Println("✅ Upload service initialized (Cloudinary)")
	}

	// Initialize handlers - Khởi tạo handler
	h := &routes.Handlers{
		Auth:         handlers.NewAuthHandler(authService, cfg),
		Story:        handlers.NewStoryHandler(storyService),
		Chapter:      handlers.NewChapterHandler(chapterService),
		Bookmark:     handlers.NewBookmarkHandler(bookmarkService),
		Comment:      handlers.NewCommentHandler(commentService),
		Notification: handlers.NewNotificationHandler(notificationService),
		Upload:       uploadHandler,
		CSRF:         handlers.NewCSRFHandler(cfg),
	}

	// Setup Gin router - Setup router cho Gin
	r := gin.New()
	r.Use(gin.Recovery())

	// Setup routes
	routes.SetupRoutes(r, cfg, h)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Không thể start server:", err)
	}
}
