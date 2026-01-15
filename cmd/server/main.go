package main

import (
	"log"
	"time"

	"nekozanedex/internal/centrifugo"
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
// @host      localhost:9091
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
		&models.CommentLike{},
		&models.CommentReport{},
	); err != nil {
		log.Fatal("Không thể migrate database:", err)
	}
	log.Println("Database Đã Migrate Thành Công")

	// One-time migration: Generate tag_name for existing users
	var usersWithoutTagName []models.User
	if err := db.Where("tag_name IS NULL OR tag_name = ''").Find(&usersWithoutTagName).Error; err == nil && len(usersWithoutTagName) > 0 {
		log.Printf("🔄 Migrating tag_name for %d users...", len(usersWithoutTagName))
		for _, user := range usersWithoutTagName {
			baseTagName := models.GenerateTagName(user.Username)
			uniqueTagName := models.GenerateUniqueTagName(db, baseTagName, user.ID)
			if err := db.Model(&user).Update("tag_name", uniqueTagName).Error; err != nil {
				log.Printf("❌ Failed to update tag_name for user %s: %v", user.Username, err)
			}
		}
		log.Println("✅ Tag name migration complete")
	}

	// Initialize repositories - Khởi tạo repository
	userRepo := repositories.NewUserRepository(db)
	storyRepo := repositories.NewStoryRepository(db)
	chapterRepo := repositories.NewChapterRepository(db)
	genreRepo := repositories.NewGenreRepository(db)
	bookmarkRepo := repositories.NewBookmarkRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	commentLikeRepo := repositories.NewCommentLikeRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	storyViewRepo := repositories.NewStoryViewRepository(db) // Fair view counting
	readingHistoryRepo := repositories.NewReadingHistoryRepository(db)
	userSettingsRepo := repositories.NewUserSettingsRepository(db)
	commentReportRepo := repositories.NewCommentReportRepository(db)

	// Init Centrifugo client
	centrifugoClient := centrifugo.NewClient(
		cfg.Centrifugo.URL,
		cfg.Centrifugo.APIKey,
		cfg.Centrifugo.SecretKey,
	)
	log.Printf("[Centrifugo] Client initialized with URL: %s", cfg.Centrifugo.URL)

	// Start background cleanup job for refresh tokens
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		
		// Run once at startup
		if err := refreshTokenRepo.DeleteExpired(); err != nil {
			log.Printf("❌ Failed to cleanup expired tokens: %v", err)
		}

		for range ticker.C {
			if err := refreshTokenRepo.DeleteExpired(); err != nil {
				log.Printf("❌ Failed to cleanup expired tokens: %v", err)
			} else {
				log.Println("🧹 Cleaned up expired and old revoked tokens")
			}
		}
	}()

	// Initialize services - Khởi tạo service
	authService := services.NewAuthService(userRepo, refreshTokenRepo, cfg) // Cập nhật với refreshTokenRepo

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

	// Pass uploadService to storyService for old cover image deletion
	storyService := services.NewStoryService(storyRepo, genreRepo, storyViewRepo, uploadService)
	genreService := services.NewGenreService(genreRepo)
	chapterService := services.NewChapterService(chapterRepo, storyRepo)
	bookmarkService := services.NewBookmarkService(bookmarkRepo, storyRepo)
	commentService := services.NewCommentService(commentRepo, storyRepo, chapterRepo)
	notificationService := services.NewNotificationService(notificationRepo, centrifugoClient)
	commentReportService := services.NewCommentReportService(commentReportRepo)

	// Start background job for scheduled chapter publishing
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		// Run once at startup
		if count, err := chapterService.PublishScheduledChapters(); err != nil {
			log.Printf("❌ Failed to publish scheduled chapters: %v", err)
		} else if count > 0 {
			log.Printf("📅 Published %d scheduled chapter(s) at startup", count)
		}

		for range ticker.C {
			if count, err := chapterService.PublishScheduledChapters(); err != nil {
				log.Printf("❌ Failed to publish scheduled chapters: %v", err)
			} else if count > 0 {
				log.Printf("📅 Auto-published %d scheduled chapter(s)", count)
			}
		}
	}()

	// Initialize handlers - Khởi tạo handler
	h := &routes.Handlers{
		Auth:           handlers.NewAuthHandler(authService, uploadService, cfg),
		Story:          handlers.NewStoryHandler(storyService),
		Chapter:        handlers.NewChapterHandler(chapterService),
		Genre:          handlers.NewGenreHandler(genreService),
		Bookmark:       handlers.NewBookmarkHandler(bookmarkService),
		Comment:        handlers.NewCommentHandler(commentService, notificationService, userRepo, storyRepo, commentLikeRepo, centrifugoClient, commentReportService),
		Notification:   handlers.NewNotificationHandler(notificationService),
		Upload:         uploadHandler,
		CSRF:           handlers.NewCSRFHandler(cfg),
		User:           handlers.NewUserHandler(userRepo),
		ReadingHistory: handlers.NewReadingHistoryHandler(readingHistoryRepo),
		UserSettings:   handlers.NewUserSettingsHandler(services.NewUserSettingsService(userSettingsRepo)),
		Centrifugo:     handlers.NewCentrifugoHandler(centrifugoClient),
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
