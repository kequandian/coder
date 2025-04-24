package server

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"coder/app"
	"coder/internal/agent"
	"coder/internal/handler"
)

// Server represents the HTTP server
type Server struct {
	handler     *handler.Handler
	ginEngine   *gin.Engine
	staticFiles fs.FS
	agent       *agent.Agent
}

// New creates a new server
func New(ctx context.Context, staticFiles fs.FS) (*Server, error) {
	// Create agent
	agent, err := agent.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Create handler
	handler := handler.New(agent)

	// Create Gin engine
	ginEngine := gin.Default()

	// Set up CORS if enabled
	if app.Config.Server.EnableCORS {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = []string{"*"}
		corsConfig.AllowMethods = []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}
		corsConfig.AllowHeaders = []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		}
		corsConfig.ExposeHeaders = []string{"Content-Length"}
		corsConfig.AllowCredentials = true
		ginEngine.Use(cors.New(corsConfig))
	}

	return &Server{
		handler:     handler,
		ginEngine:   ginEngine,
		staticFiles: staticFiles,
		agent:       agent,
	}, nil
}

// SetupRoutes sets up the server routes
func (s *Server) SetupRoutes() {
	// API routes
	api := s.ginEngine.Group("/api")
	{
		api.GET("/health", s.handler.HandleHealthCheck)
	}

	// OpenAI-compatible chat completions endpoint
	s.ginEngine.POST("/v1/chat/completions", gin.WrapF(s.handler.HandleChatCompletion))

	// Serve static files with a specific prefix
	s.ginEngine.StaticFS("/admin/", http.FS(s.staticFiles))

	// Handle root route specifically
	s.ginEngine.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(s.staticFiles))
	})

	// Catch all remaining routes to serve the SPA index
	s.ginEngine.NoRoute(func(c *gin.Context) {
		// Skip API and OpenAI routes
		if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
			strings.HasPrefix(c.Request.URL.Path, "/v1/") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// Check if this is a static file request that should be served from /static/
		// Strip the leading slash and check if file exists
		filePath := c.Request.URL.Path
		if filePath != "" && filePath[0] == '/' {
			filePath = filePath[1:]
		}

		// Try to serve the file from static files
		if _, err := fs.Stat(s.staticFiles, filePath); err == nil {
			// File exists, redirect to /static/ path
			c.Redirect(http.StatusTemporaryRedirect, "/static/"+filePath)
			return
		}

		// Default SPA handling - serve index.html
		c.FileFromFS("index.html", http.FS(s.staticFiles))
	})
}

// Start starts the server
func (s *Server) Start() {
	addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Access web interface at http://%s:%d",
		func() string {
			if app.Config.Server.Host == "0.0.0.0" || app.Config.Server.Host == "::" {
				return "localhost"
			}
			return app.Config.Server.Host
		}(),
		app.Config.Server.Port)

	// Set up graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: s.ginEngine,
	}

	// Run server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), app.Config.Server.ShutdownTimeout)
	defer cancel()

	// Shut down server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Clean up resources
	s.Cleanup()

	log.Println("Server exited")
}

// Cleanup cleans up resources
func (s *Server) Cleanup() {
	// Clean up agent resources
	if s.agent != nil {
		s.agent.Close()
	}
}
