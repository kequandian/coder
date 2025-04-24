package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"coder/app"
	"coder/config"
	"coder/internal/server"
)

//go:embed static
var staticFiles embed.FS

func main() {
	// 加载配置
	app.Init()
	// 设置日志
	setupLogging(app.Config)

	// 设置上下文
	ctx := context.Background()

	// 设置静态文件
	staticAssets, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Failed to set up static assets: %v", err)
	}

	// 创建服务器
	s, err := server.New(ctx, staticAssets)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 设置路由
	s.SetupRoutes()

	// 启动服务器
	s.Start()
}

// setupLogging 配置日志系统
func setupLogging(cfg *config.Config) {
	if cfg.LogPath == "" {
		return
	}

	// 确保日志目录存在
	logDir := filepath.Dir(cfg.LogPath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Printf("Failed to create log directory %s: %v", logDir, err)
			return
		}
	}

	// 打开日志文件
	logFile, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file %s: %v", cfg.LogPath, err)
		return
	}

	// 设置日志输出
	log.SetOutput(logFile)
	log.Printf("Logging to %s", cfg.LogPath)
}
