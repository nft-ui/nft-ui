package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := log.New(os.Stdout, "[nft-ui] ", log.LstdFlags)

	// Initialize NFT manager
	nftMgr := NewNFTManager(cfg)

	// Initialize handler
	handler := NewHandler(nftMgr, cfg, logger)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// API routes
	api := e.Group("/api/v1")
	api.Use(BasicAuthMiddleware(cfg))
	api.Use(ReadOnlyMiddleware(cfg))
	api.Use(AuditLogMiddleware(logger))

	// Register API endpoints
	api.GET("/quotas", handler.ListQuotas)
	api.POST("/quotas/:id/reset", handler.ResetQuota)
	api.POST("/quotas/batch-reset", handler.BatchResetQuotas)
	api.PUT("/quotas/:id", handler.ModifyQuota)
	api.POST("/quotas", handler.AddQuota)
	api.DELETE("/quotas/:id", handler.DeleteQuota)

	// Port management endpoints
	api.POST("/ports", handler.AddPort)
	api.DELETE("/ports/:handle", handler.DeletePort)

	// Serve frontend
	setupFrontend(e)

	// Start server
	logger.Printf("Starting server on %s (read_only=%v, auth=%v)",
		cfg.ListenAddr, cfg.ReadOnly, cfg.AuthEnabled())
	e.Logger.Fatal(e.Start(cfg.ListenAddr))
}

// setupFrontend configures the embedded frontend serving
func setupFrontend(e *echo.Echo) {
	// Get the frontend/dist subdirectory
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		// Frontend not built yet, serve a placeholder
		e.GET("/*", func(c echo.Context) error {
			return c.HTML(http.StatusOK, `<!DOCTYPE html>
<html>
<head><title>nft-ui</title></head>
<body>
<h1>nft-ui</h1>
<p>Frontend not built. Run <code>cd frontend && npm install && npm run build</code></p>
<p>API available at <a href="/api/v1/quotas">/api/v1/quotas</a></p>
</body>
</html>`)
		})
		return
	}

	// Create file server for static files
	fileServer := http.FileServer(http.FS(distFS))

	// Serve static files and handle SPA routing
	e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Try to open the file
		f, err := distFS.Open(path[1:]) // Remove leading /
		if err != nil {
			// File not found, serve index.html for SPA routing
			r.URL.Path = "/index.html"
		} else {
			f.Close()
		}

		fileServer.ServeHTTP(w, r)
	})))
}
