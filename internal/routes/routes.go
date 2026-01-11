package routes

import (
	"nekozanedex/internal/config"
	"nekozanedex/internal/handlers"
	"nekozanedex/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Auth         *handlers.AuthHandler
	Story        *handlers.StoryHandler
	Chapter      *handlers.ChapterHandler
	Bookmark     *handlers.BookmarkHandler
	Comment      *handlers.CommentHandler
	Notification *handlers.NotificationHandler
	Upload       *handlers.UploadHandler
	CSRF         *handlers.CSRFHandler
}

func SetupRoutes(r *gin.Engine, cfg *config.Config, h *Handlers) {
	// Security Middleware
	r.Use(middleware.SecurityHeaders(cfg))
	r.Use(middleware.CORSMiddleware(cfg))
	r.Use(middleware.LoggerMiddleware())

	csrfCfg := middleware.DefaultCSRFConfig()
	csrfCfg.SecretKey = cfg.CSRF.SecretKey
	csrfCfg.Secure = cfg.App.IsProduction
	// Production-only security headers (HSTS)
	if cfg.App.IsProduction {
		r.Use(middleware.ProductionSecurityHeaders())
	}

	// Health check (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger API Documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes với General Rate Limiting (100 req/min)
	api := r.Group("/api")
	api.Use(middleware.GeneralRateLimiter())
	{
		// ============ AUTH ROUTES ============
		auth := api.Group("/auth")
		{
			// Stricter rate limit cho auth mutations (chống brute-force)
			auth.POST("/register", middleware.AuthRateLimiter(), h.Auth.Register)
			auth.POST("/login", middleware.AuthRateLimiter(), h.Auth.Login)
			auth.POST("/refresh", middleware.AuthRateLimiter(), h.Auth.RefreshToken)
			auth.POST("/logout", h.Auth.Logout)

			// Protected routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(cfg))
			{
				authProtected.GET("/profile", h.Auth.GetProfile)
				authProtected.PUT("/profile", h.Auth.UpdateProfile)
				authProtected.POST("/change-password", h.Auth.ChangePassword)
				authProtected.POST("/logout-all", h.Auth.LogoutAll)
				authProtected.GET("/sessions", h.Auth.GetSessions)
				authProtected.GET("/csrf-token", h.CSRF.GetCSRFToken) // CSRF token refresh
			}
		}

		// ============ STORY ROUTES (Public) ============
		stories := api.Group("/stories")
		{
			stories.GET("", h.Story.GetStories)
			stories.GET("/latest", h.Story.GetLatestStories)
			stories.GET("/hot", h.Story.GetHotStories)
			stories.GET("/random", h.Story.GetRandomStory)
			stories.GET("/search", h.Story.SearchStories)
			stories.GET("/:slug", h.Story.GetStoryBySlug)
			stories.GET("/:slug/chapters", h.Chapter.GetChaptersByStory)
			stories.GET("/:slug/chapters/:number", h.Chapter.GetChapterByNumber)
		}

		// ============ GENRE ROUTES ============
		genres := api.Group("/genres")
		{
			genres.GET("", h.Story.GetAllGenres)
			genres.GET("/:genre/stories", h.Story.GetStoriesByGenre)
		}

		// ============ COMMENT ROUTES ============
		comments := api.Group("/comments")
		comments.Use(middleware.OptionalAuthMiddleware(cfg))
		{
			comments.GET("/story/:storyId", h.Comment.GetCommentsByStory)
			comments.GET("/chapter/:chapterId", h.Comment.GetCommentsByChapter)
		}

		commentsAuth := api.Group("/comments")
		commentsAuth.Use(middleware.AuthMiddleware(cfg))
		commentsAuth.Use(middleware.RoleMiddleware("reader", "admin")) // Reader hoặc Admin
		{
			commentsAuth.POST("/:commentId/reply", h.Comment.ReplyComment)
			commentsAuth.DELETE("/:commentId", h.Comment.DeleteComment)
		}

		// Story comments (authenticated)
		api.POST("/stories/:storyId/comments", middleware.AuthMiddleware(cfg), h.Comment.CreateComment)

		// ============ BOOKMARK ROUTES (Reader + Admin) ============
		bookmarks := api.Group("/bookmarks")
		bookmarks.Use(middleware.AuthMiddleware(cfg))
		bookmarks.Use(middleware.RoleMiddleware("reader", "admin"))
		{
			bookmarks.GET("", h.Bookmark.GetMyBookmarks)
			bookmarks.POST("/:storyId", h.Bookmark.AddBookmark)
			bookmarks.DELETE("/:storyId", h.Bookmark.RemoveBookmark)
			bookmarks.GET("/:storyId/check", h.Bookmark.CheckBookmark)
		}

		// ============ NOTIFICATION ROUTES (Reader + Admin) ============
		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(cfg))
		notifications.Use(middleware.RoleMiddleware("reader", "admin"))
		{
			notifications.GET("", h.Notification.GetMyNotifications)
			notifications.GET("/unread-count", h.Notification.GetUnreadCount)
			notifications.POST("/:id/read", h.Notification.MarkAsRead)
			notifications.POST("/read-all", h.Notification.MarkAllAsRead)
		}

		// ============ ADMIN ROUTES (Admin Only) ============
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		admin.Use(middleware.CSRFMiddleware(csrfCfg))
		admin.Use(middleware.RoleMiddleware("admin")) // Chỉ Admin
		{
			// Admin Stories
			adminStories := admin.Group("/stories")
			{
				adminStories.GET("", h.Story.GetAllStoriesAdmin)
				adminStories.GET("/:id", h.Story.GetStoryByID)
				adminStories.POST("", h.Story.CreateStory)
				adminStories.PUT("/:id", h.Story.UpdateStory)
				adminStories.DELETE("/:id", h.Story.DeleteStory)

				// Admin Chapters (nested under stories)
				adminStories.GET("/:id/chapters", h.Chapter.GetChaptersByStoryAdmin)
				adminStories.POST("/:id/chapters", h.Chapter.CreateChapter)
				adminStories.POST("/:id/chapters/bulk", h.Chapter.BulkImportChapters)
			}

			// Admin Chapters
			adminChapters := admin.Group("/chapters")
			{
				adminChapters.GET("/:id", h.Chapter.GetChapterByID)
				adminChapters.PUT("/:id", h.Chapter.UpdateChapter)
				adminChapters.DELETE("/:id", h.Chapter.DeleteChapter)
				adminChapters.POST("/:id/publish", h.Chapter.PublishChapter)
				adminChapters.POST("/:id/schedule", h.Chapter.ScheduleChapter)
			}

			// Admin Media (Cloudinary uploads for stories/chapters)
			if h.Upload != nil {
				adminMedia := admin.Group("/media")
				{
					adminMedia.POST("", h.Upload.UploadSingleImage)
					adminMedia.POST("/chapter", h.Upload.UploadChapterImages)
					adminMedia.DELETE("", h.Upload.DeleteImage)
				}
			}
		}

		// ============ USER ROUTES (Authenticated Users) ============
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg))
		{
			// User avatar upload (any authenticated user)
			if h.Upload != nil {
				users.POST("/upload-avatar", h.Upload.UploadAvatar)
			}
		}
	}
}
