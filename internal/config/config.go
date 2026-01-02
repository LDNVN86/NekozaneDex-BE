package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App			AppConfig
	Server		ServerConfig
	Database	DatabaseConfig
	Jwt			JwtConfig
	Cookie		CookieConfig
	Centrifugo	CentrifugoConfig	
}

type AppConfig struct {
	Env		string
	IsProduction	bool
}

type ServerConfig struct {
	Port		string
	GinMode		string
}

type DatabaseConfig struct {
	Host		string
	Port		string
	User		string
	Password	string
	DBName		string
}

type JwtConfig struct {
	AccessSecret		string
	RefreshSecret		string
	AccessExpireMinutes	int
	RefreshExpireDays	int
}

type CookieConfig struct {
	Domain		string
	Secure		bool
	HttpOnly	bool
	SameSite	string
	Path		string
	MaxAge		int
}

type CentrifugoConfig struct {
	Url		string
	APIKey		string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	env := getEnv("APP_ENV", "development")
	isProduction := env == "production"
	cookieSecure := isProduction

	accessExpire,_ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRE_MINUTES","30"))
	refreshExpire,_ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRE_DAYS","7"))
	cookieMaxAge,_ := strconv.Atoi(getEnv("JWT_COOKIE_MAX_AGE","604800"))

	return &Config{
		App: AppConfig{
			Env:      env,
			IsProduction: isProduction,
		},
		Server: ServerConfig{
			Port: getEnv("PORT","8080"),
			GinMode: getEnv("GIN_MODE","debug"),
		},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST","localhost"),
			Port: getEnv("DB_PORT","5432"),
			User: getEnv("DB_USER","postgres"),
			Password: getEnv("DB_PASSWORD",""),
			DBName: getEnv("DB_NAME","nekozanedex"),
		},
		Jwt: JwtConfig{
			AccessSecret: getEnv("JWT_ACCESS_SECRET","your-super-secret-key-change-this-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET","your-super-secret-key-change-this-in-production"),
			AccessExpireMinutes: accessExpire,
			RefreshExpireDays: refreshExpire,
		},
		Centrifugo: CentrifugoConfig{
			Url: getEnv("CENTRIFUGO_URL","http://localhost:8000"),
			APIKey: getEnv("CENTRIFUGO_API_KEY","your-centrifugo-api-key"),
		},
		Cookie: CookieConfig{
			Domain: getEnv("JWT_COOKIE_DOMAIN","localhost"),
			Secure: cookieSecure,
			HttpOnly: true,
			SameSite: getEnv("JWT_COOKIE_SAME_SITE","lax"),
			Path: "/",
			MaxAge: cookieMaxAge,
		},
	}, nil
}

//Get env variable

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

//Helper Methods
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

func (c *Config) IsStaging() bool {
	return c.App.Env == "staging"
}