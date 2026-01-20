package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// BasicAuthMiddleware creates Echo middleware for HTTP Basic Auth
func BasicAuthMiddleware(cfg *Config) echo.MiddlewareFunc {
	if !cfg.AuthEnabled() {
		// No auth configured, skip
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Use constant-time comparison to prevent timing attacks
		userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(cfg.AuthUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(cfg.AuthPassword)) == 1
		return userMatch && passMatch, nil
	})
}

// ReadOnlyMiddleware blocks write operations when in read-only mode
func ReadOnlyMiddleware(cfg *Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.ReadOnly && c.Request().Method != http.MethodGet {
				return c.JSON(http.StatusForbidden, APIResponse{
					Success: false,
					Error:   "Server is in read-only mode",
				})
			}
			return next(c)
		}
	}
}

// AuditLogMiddleware logs all modification operations
func AuditLogMiddleware(logger *log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only log non-GET requests (modifications)
			if c.Request().Method != http.MethodGet {
				user, _, _ := c.Request().BasicAuth()
				if user == "" {
					user = "anonymous"
				}
				logger.Printf("[AUDIT] %s %s %s from %s by user=%s",
					time.Now().Format(time.RFC3339),
					c.Request().Method,
					c.Request().URL.Path,
					c.RealIP(),
					user,
				)
			}
			return next(c)
		}
	}
}

// RateLimitMiddleware provides simple in-memory rate limiting per IP
func RateLimitMiddleware(limit int, window time.Duration) echo.MiddlewareFunc {
	var mu sync.Mutex
	requests := make(map[string][]time.Time)

	// Cleanup old entries periodically
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, times := range requests {
				var valid []time.Time
				for _, t := range times {
					if now.Sub(t) < window {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(requests, ip)
				} else {
					requests[ip] = valid
				}
			}
			mu.Unlock()
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			now := time.Now()

			mu.Lock()
			// Clean old entries for this IP
			var validRequests []time.Time
			for _, t := range requests[ip] {
				if now.Sub(t) < window {
					validRequests = append(validRequests, t)
				}
			}

			if len(validRequests) >= limit {
				mu.Unlock()
				return c.JSON(http.StatusTooManyRequests, APIResponse{
					Success: false,
					Error:   "Rate limit exceeded. Please try again later.",
				})
			}

			requests[ip] = append(validRequests, now)
			mu.Unlock()

			return next(c)
		}
	}
}
