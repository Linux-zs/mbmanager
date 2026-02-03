package api

import (
	"mbmanager/internal/api/handler"
	"mbmanager/internal/api/middleware"
	"mbmanager/internal/service"

	"github.com/gin-gonic/gin"
)

var schedulerService *service.SchedulerService

// SetSchedulerService 设置调度器服务
func SetSchedulerService(svc *service.SchedulerService) {
	schedulerService = svc
	handler.SetSchedulerService(svc)
}

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(middleware.CORS())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 认证
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.POST("/logout", handler.Logout)
		}

		// 需要认证的路由
		authorized := v1.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			// 仪表盘
			authorized.GET("/dashboard/stats", handler.GetDashboardStats)

			// 主机管理
			hosts := authorized.Group("/hosts")
			{
				hosts.GET("", handler.GetHosts)
				hosts.POST("", handler.CreateHost)
				hosts.GET("/:id", handler.GetHost)
				hosts.PUT("/:id", handler.UpdateHost)
				hosts.DELETE("/:id", handler.DeleteHost)
				hosts.POST("/:id/test", handler.TestHostConnection)
			}

			// 任务管理
			tasks := authorized.Group("/tasks")
			{
				tasks.GET("", handler.GetTasks)
				tasks.POST("", handler.CreateTask)
				tasks.GET("/:id", handler.GetTask)
				tasks.PUT("/:id", handler.UpdateTask)
				tasks.DELETE("/:id", handler.DeleteTask)
				tasks.POST("/:id/run", handler.RunTask)
				tasks.GET("/:id/logs", handler.GetTaskLogs)
			}

			// 存储管理
			storages := authorized.Group("/storages")
			{
				storages.GET("", handler.GetStorages)
				storages.POST("", handler.CreateStorage)
				storages.GET("/:id", handler.GetStorage)
				storages.PUT("/:id", handler.UpdateStorage)
				storages.DELETE("/:id", handler.DeleteStorage)
				storages.POST("/:id/test", handler.TestStorageConnection)
			storages.GET("/:id/diskspace", handler.GetStorageDiskSpace)
			}

			// 通知管理
			notifications := authorized.Group("/notifications")
			{
				notifications.GET("", handler.GetNotifications)
				notifications.POST("", handler.CreateNotification)
				notifications.GET("/:id", handler.GetNotification)
				notifications.PUT("/:id", handler.UpdateNotification)
				notifications.DELETE("/:id", handler.DeleteNotification)
				notifications.POST("/:id/test", handler.TestNotification)
			}

			// 备份日志
			logs := authorized.Group("/logs")
			{
				logs.GET("", handler.GetLogs)
				logs.GET("/:id", handler.GetLog)
				logs.DELETE("/:id", handler.DeleteLog)
			}

			// 备份文件管理
			backups := authorized.Group("/backups")
			{
				backups.DELETE("/:id", handler.DeleteBackup)
				backups.GET("/:id/download", handler.DownloadBackup)
			}

			// 用户管理
			users := authorized.Group("/users")
			{
				users.GET("", handler.GetUsers)
				users.POST("", handler.CreateUser)
				users.GET("/:id", handler.GetUser)
				users.PUT("/:id", handler.UpdateUser)
				users.DELETE("/:id", handler.DeleteUser)
			}
		}
	}

	// 静态文件
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("./web/templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	return router
}
