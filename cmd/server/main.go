package main

import (
	"context"
	"mbmanager/internal/api"
	"mbmanager/internal/config"
	"mbmanager/internal/database"
	"mbmanager/internal/logger"
	"mbmanager/internal/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

var schedulerService *service.SchedulerService

func main() {
	// 初始化日志系统
	if err := logger.InitLogger("./logs"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.CloseLogger()

	logger.Info("Starting mbmanager server...")

	// 加载配置
	cfg := config.LoadConfig()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := database.InitDB(cfg.Database.Path); err != nil {
		logger.Error("Failed to initialize database: %v", err)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	logger.Info("Database initialized successfully")

	// 创建备份目录
	if err := os.MkdirAll(cfg.Backup.BasePath, 0755); err != nil {
		logger.Error("Failed to create backup directory: %v", err)
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// 创建临时目录
	if err := os.MkdirAll("./data/tmp", 0755); err != nil {
		logger.Error("Failed to create temp directory: %v", err)
		log.Fatalf("Failed to create temp directory: %v", err)
	}

	// 初始化调度服务
	backupSvc := service.NewBackupService()
	var err error
	schedulerService, err = service.NewSchedulerService(backupSvc)
	if err != nil {
		logger.Error("Failed to create scheduler service: %v", err)
		log.Fatalf("Failed to create scheduler service: %v", err)
	}

	// 启动调度器
	ctx := context.Background()
	if err := schedulerService.Start(ctx); err != nil {
		logger.Error("Failed to start scheduler: %v", err)
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	logger.Info("Scheduler started successfully")

	// 创建Gin路由（传递调度器服务）
	router := api.SetupRouter()

	// 设置全局调度器服务（供API使用）
	api.SetSchedulerService(schedulerService)

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down gracefully...")
		log.Println("Shutting down gracefully...")
		if err := schedulerService.Stop(); err != nil {
			logger.Error("Error stopping scheduler: %v", err)
			log.Printf("Error stopping scheduler: %v", err)
		}
		os.Exit(0)
	}()

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Server starting on %s", addr)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		logger.Error("Failed to start server: %v", err)
		log.Fatalf("Failed to start server: %v", err)
	}
}
