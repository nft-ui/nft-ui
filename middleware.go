package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// AuthFailureTracker tracks authentication failures per IP
type AuthFailureTracker struct {
	mu            sync.Mutex
	failures      map[string]int       // IP -> failure count
	lockUntil     map[string]time.Time // IP -> unlock time
	maxAttempts   int                  // Max failed attempts before lock
	lockDuration  time.Duration        // How long to lock an IP
	windowDuration time.Duration       // Time window for counting failures
	lastFailure   map[string]time.Time // IP -> last failure time
}

// NewAuthFailureTracker creates a new tracker with default settings
func NewAuthFailureTracker() *AuthFailureTracker {
	tracker := &AuthFailureTracker{
		failures:       make(map[string]int),
		lockUntil:      make(map[string]time.Time),
		lastFailure:    make(map[string]time.Time),
		maxAttempts:    5,                // 5 failed attempts
		lockDuration:   15 * time.Minute, // Lock for 15 minutes
		windowDuration: 5 * time.Minute,  // Reset counter after 5 minutes of no failures
	}

	// Cleanup goroutine
	go tracker.cleanup()

	return tracker
}

// cleanup removes expired entries periodically
func (t *AuthFailureTracker) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		t.mu.Lock()
		now := time.Now()
		for ip, unlockTime := range t.lockUntil {
			if now.After(unlockTime) {
				delete(t.lockUntil, ip)
				delete(t.failures, ip)
				delete(t.lastFailure, ip)
			}
		}
		// Reset failure counts for IPs that haven't failed in windowDuration
		for ip, lastFail := range t.lastFailure {
			if now.Sub(lastFail) > t.windowDuration {
				delete(t.failures, ip)
				delete(t.lastFailure, ip)
			}
		}
		t.mu.Unlock()
	}
}

// IsLocked checks if an IP is currently locked
func (t *AuthFailureTracker) IsLocked(ip string) (bool, time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if unlockTime, exists := t.lockUntil[ip]; exists {
		if time.Now().Before(unlockTime) {
			return true, unlockTime
		}
		// Lock expired, clean up
		delete(t.lockUntil, ip)
		delete(t.failures, ip)
		delete(t.lastFailure, ip)
	}
	return false, time.Time{}
}

// RecordFailure records an authentication failure
func (t *AuthFailureTracker) RecordFailure(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.lastFailure[ip] = now

	// Increment failure count
	t.failures[ip]++

	// Check if we should lock this IP
	if t.failures[ip] >= t.maxAttempts {
		t.lockUntil[ip] = now.Add(t.lockDuration)
	}
}

// RecordSuccess clears failure records for an IP
func (t *AuthFailureTracker) RecordSuccess(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.failures, ip)
	delete(t.lockUntil, ip)
	delete(t.lastFailure, ip)
}

// GetRemainingAttempts returns how many attempts are left before lock
func (t *AuthFailureTracker) GetRemainingAttempts(ip string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.maxAttempts - t.failures[ip]
}

var globalAuthTracker = NewAuthFailureTracker()

// BasicAuthMiddleware creates Echo middleware for HTTP Basic Auth with brute-force protection
func BasicAuthMiddleware(cfg *Config) echo.MiddlewareFunc {
	if !cfg.AuthEnabled() {
		// No auth configured, skip
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			// Check if IP is locked
			if locked, unlockTime := globalAuthTracker.IsLocked(ip); locked {
				remainingTime := time.Until(unlockTime).Round(time.Second)
				return c.JSON(http.StatusTooManyRequests, APIResponse{
					Success: false,
					Error:   "Too many failed authentication attempts. Please try again in " + remainingTime.String(),
				})
			}

			// Extract credentials
			username, password, ok := c.Request().BasicAuth()
			if !ok {
				c.Response().Header().Set("WWW-Authenticate", `Basic realm="nft-ui"`)
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			// Validate credentials using constant-time comparison
			userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(cfg.AuthUser)) == 1
			passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(cfg.AuthPassword)) == 1

			if userMatch && passMatch {
				// Success - clear any failure records
				globalAuthTracker.RecordSuccess(ip)
				return next(c)
			}

			// Failure - record it
			globalAuthTracker.RecordFailure(ip)
			remaining := globalAuthTracker.GetRemainingAttempts(ip)

			if remaining > 0 {
				c.Response().Header().Set("WWW-Authenticate", `Basic realm="nft-ui"`)
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
			} else {
				return c.JSON(http.StatusTooManyRequests, APIResponse{
					Success: false,
					Error:   "Too many failed attempts. Account locked for 15 minutes.",
				})
			}
		}
	}
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
