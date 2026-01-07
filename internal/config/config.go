package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	Server     ServerConfig
	Database   DatabaseConfig
	Jwt        JwtConfig
	Cookie     CookieConfig
	Security   SecurityConfig
	Centrifugo CentrifugoConfig
	Cloudinary CloudinaryConfig
	CSRF       CSRFConfig
}

type AppConfig struct {
	Env          string
	IsProduction bool
}

type SecurityConfig struct {
	FrameAncestors string
}

type CSRFConfig struct {
	SecretKey 		string
	Secure			bool
	CookieDomain	[]string
	CookiePath		[]string
	HeaderName		[]string
	CookieName		string
	ExcludePaths	[]string

}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JwtConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpireMinutes int
	RefreshExpireDays   int
}

type CookieConfig struct {
	Domain   string
	Secure   bool
	HttpOnly bool
	SameSite string
	Path     string
	MaxAge   int
}

type CentrifugoConfig struct {
	Url    string
	APIKey string
}

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
}

// LoadConfig - Load cấu hình từ file .env
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Không tìm thấy file .env, sử dụng biến môi trường")
	}

	env := getEnv("APP_ENV", "development")
	isProduction := env == "production"
	cookieSecure := isProduction
	cookieDomain := getEnv("JWT_COOKIE_DOMAIN", "")
	if !isProduction {
		if cookieDomain == "localhost" || cookieDomain == "127.0.0.1" {
			cookieDomain = ""
		}
	}
	frameAncestors := getEnv("FRAME_ANCESTORS", "'self'")

	accessExpire, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRE_MINUTES", "30"))
	refreshExpire, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRE_DAYS", "7"))
	cookieMaxAge, _ := strconv.Atoi(getEnv("JWT_COOKIE_MAX_AGE", "604800"))

	return &Config{
		App: AppConfig{
			Env:          env,
			IsProduction: isProduction,
		},
		Server: ServerConfig{
			Port:    getEnv("PORT", "9091"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "nekozanedex"),
		},
		Jwt: JwtConfig{
			AccessSecret:        getEnv("JWT_ACCESS_SECRET", "Thay-Bang-Key-Khac-Khi-Len_Production"),
			RefreshSecret:       getEnv("JWT_REFRESH_SECRET", "Thay-Bang-Key-Khac-Khi-Len_Production"),
			AccessExpireMinutes: accessExpire,
			RefreshExpireDays:   refreshExpire,
		},
		Centrifugo: CentrifugoConfig{
			Url:    getEnv("CENTRIFUGO_URL", "http://localhost:9091"),
			APIKey: getEnv("CENTRIFUGO_API_KEY", "Thay-Bang-Key-Khac-Khi-Len_Production"),
		},
		Cookie: CookieConfig{
			Domain:   cookieDomain,
			Secure:   cookieSecure,
			HttpOnly: true,
			SameSite: func() string {
				if isProduction {
					return "strict" // Production: strict để chống CSRF
				}
				return getEnv("JWT_COOKIE_SAME_SITE", "lax") // Dev: lax cho dễ test
			}(),
			Path:   "/",
			MaxAge: cookieMaxAge,
		},
		Security: SecurityConfig{
			FrameAncestors: frameAncestors,
		},
		Cloudinary: CloudinaryConfig{
			CloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
			APIKey:    getEnv("CLOUDINARY_API_KEY", ""),
			APISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		},
		CSRF: CSRFConfig{
			SecretKey: getEnv("CSRF_SECRET_KEY", "default-dev-secret-key-change-this"),
		},
	}, nil
}

//Lấy Biến Môi Trường - Get Environment Variable
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper Method - Phương Thức Hỗ Trợ
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

func (c *Config) IsStaging() bool {
	return c.App.Env == "staging"
}
