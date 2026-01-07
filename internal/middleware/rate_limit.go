package middleware

import (
	"strconv"
	"sync"
	"time"

	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter - In-memory rate limiter - Tối ưu hóa hiệu suất
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int           // Số requests tối đa
	window   time.Duration // Thời gian window
}

type clientInfo struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter - Tạo rate limiter mới
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
	}

	// Cleanup routine - xóa các entries hết hạn
	go rl.cleanup()

	return rl
}

// Middleware - Rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()

		// Lấy hoặc tạo client info
		client, exists := rl.requests[ip]
		now := time.Now()

		if !exists || now.After(client.resetTime) {
			// Client mới hoặc đã reset
			rl.requests[ip] = &clientInfo{
				count:     1,
				resetTime: now.Add(rl.window),
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Kiểm tra rate limit
		if client.count >= rl.limit {
			rl.mu.Unlock()

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", client.resetTime.Format(time.RFC3339))
			c.Header("Retry-After", strconv.Itoa(int(client.resetTime.Sub(now).Seconds())))

			response.TooManyRequests(c, "Quá nhiều requests, vui lòng thử lại sau")
			c.Abort()
			return
		}

		// Tăng count
		client.count++
		remaining := rl.limit - client.count
		rl.mu.Unlock()

		// Set headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", client.resetTime.Format(time.RFC3339))

		c.Next()
	}
}

// cleanup - Xóa các entries hết hạn (chạy background)
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.requests {
			if now.After(client.resetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// ============ PRE-CONFIGURED RATE LIMITERS ============

// GeneralRateLimiter - 100 requests/phút cho API thông thường
func GeneralRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(100, 1*time.Minute)
	return limiter.Middleware()
}

// AuthRateLimiter - 10 requests/phút cho auth endpoints (chống brute-force)
func AuthRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(10, 1*time.Minute)
	return limiter.Middleware()
}

// StrictRateLimiter - 5 requests/phút cho sensitive endpoints
func StrictRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(5, 1*time.Minute)
	return limiter.Middleware()
}

// HealthCheckSkip - Skip health check từ rate limiting
func HealthCheckSkip(c *gin.Context) bool {
	return c.Request.URL.Path == "/health"
}
