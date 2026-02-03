package handler

import (
	"context"
	"mbmanager/internal/database"
	"mbmanager/internal/logger"
	"mbmanager/internal/model"
	"mbmanager/internal/service"
	"mbmanager/internal/storage"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var schedulerService *service.SchedulerService

// SetSchedulerService 设置调度器服务
func SetSchedulerService(svc *service.SchedulerService) {
	schedulerService = svc
}

// GetDashboardStats 获取仪表盘统计
func GetDashboardStats(c *gin.Context) {
	var taskCount int64
	var hostCount int64
	var successCount int64
	var failedCount int64

	database.DB.Model(&model.Task{}).Count(&taskCount)
	database.DB.Model(&model.Host{}).Count(&hostCount)
	database.DB.Model(&model.BackupLog{}).Where("status = ?", "success").Count(&successCount)
	database.DB.Model(&model.BackupLog{}).Where("status = ?", "failed").Count(&failedCount)

	c.JSON(http.StatusOK, gin.H{
		"task_count":    taskCount,
		"host_count":    hostCount,
		"success_count": successCount,
		"failed_count":  failedCount,
	})
}

// GetHosts 获取主机列表
func GetHosts(c *gin.Context) {
	var hosts []model.Host
	if err := database.DB.Find(&hosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hosts)
}

// CreateHost 创建主机
func CreateHost(c *gin.Context) {
	var host model.Host
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试连接并获取MySQL版本
	hostSvc := service.NewHostService()
	if err := hostSvc.TestConnection(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Connection test failed: %v", err)})
		return
	}

	if err := database.DB.Create(&host).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, host)
}

// GetHost 获取主机详情
func GetHost(c *gin.Context) {
	id := c.Param("id")
	var host model.Host
	if err := database.DB.First(&host, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		return
	}
	c.JSON(http.StatusOK, host)
}

// UpdateHost 更新主机
func UpdateHost(c *gin.Context) {
	id := c.Param("id")
	var host model.Host
	if err := database.DB.First(&host, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		return
	}

	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试连接并获取MySQL版本
	hostSvc := service.NewHostService()
	if err := hostSvc.TestConnection(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Connection test failed: %v", err)})
		return
	}

	if err := database.DB.Save(&host).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, host)
}

// DeleteHost 删除主机
func DeleteHost(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&model.Host{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Host deleted successfully"})
}

// TestHostConnection 测试主机连接
func TestHostConnection(c *gin.Context) {
	id := c.Param("id")
	var host model.Host
	if err := database.DB.First(&host, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		return
	}

	// 测试连接
	hostSvc := service.NewHostService()
	if err := hostSvc.TestConnection(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	// 保存MySQL版本
	if host.MySQLVersion != "" {
		database.DB.Model(&host).Update("mysql_version", host.MySQLVersion)
	}

	// 获取数据库列表
	databases, err := hostSvc.GetDatabases(&host)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection successful", "databases": []string{}, "version": host.MySQLVersion})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection successful", "databases": databases, "version": host.MySQLVersion})
}

// GetTasks 获取任务列表
func GetTasks(c *gin.Context) {
	var tasks []model.Task
	if err := database.DB.Preload("Host").Preload("Storage").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 为每个任务添加最后备份信息
	type TaskWithLastBackup struct {
		model.Task
		LastBackup *model.BackupLog `json:"last_backup"`
	}

	var result []TaskWithLastBackup
	for _, task := range tasks {
		var lastBackup model.BackupLog
		err := database.DB.Where("task_id = ?", task.ID).
			Order("created_at DESC").
			First(&lastBackup).Error

		taskWithBackup := TaskWithLastBackup{
			Task: task,
		}
		if err == nil {
			taskWithBackup.LastBackup = &lastBackup
		}
		result = append(result, taskWithBackup)
	}

	c.JSON(http.StatusOK, result)
}

// CreateTask 创建任务
func CreateTask(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果任务启用，添加到调度器
	if task.Status == 1 && schedulerService != nil {
		if err := schedulerService.AddTask(&task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Task created but failed to schedule: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask 获取任务详情
func GetTask(c *gin.Context) {
	id := c.Param("id")
	var task model.Task
	if err := database.DB.Preload("Host").Preload("Storage").First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// UpdateTask 更新任务
func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var task model.Task
	if err := database.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// 绑定更新数据到新的结构体
	var updateData model.Task
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保留ID，使用Updates更新
	updateData.ID = task.ID
	if err := database.DB.Model(&task).Updates(&updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载任务以获取最新数据
	if err := database.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload task"})
		return
	}

	// 更新调度器
	if schedulerService != nil {
		if task.Status == 1 {
			// 任务启用，更新调度
			if err := schedulerService.UpdateTask(&task); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Task updated but failed to reschedule: " + err.Error()})
				return
			}
		} else {
			// 任务禁用，从调度器移除
			schedulerService.RemoveTask(task.ID)
		}
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask 删除任务
func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// 解析ID
	taskID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// 从调度器移除
	if schedulerService != nil {
		schedulerService.RemoveTask(uint(taskID))
	}

	// 删除任务
	if err := database.DB.Delete(&model.Task{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// RunTask 立即执行任务
func RunTask(c *gin.Context) {
	id := c.Param("id")

	// 解析ID
	taskID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// 执行任务
	if schedulerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scheduler service not available"})
		return
	}

	if err := schedulerService.RunTaskNow(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task started successfully"})
}

// GetTaskLogs 获取任务的备份日志
func GetTaskLogs(c *gin.Context) {
	id := c.Param("id")
	var logs []model.BackupLog

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	query := database.DB.Model(&model.BackupLog{}).Where("task_id = ?", id)
	query.Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStorages 获取存储列表
func GetStorages(c *gin.Context) {
	var storages []model.Storage
	if err := database.DB.Find(&storages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, storages)
}

// CreateStorage 创建存储
func CreateStorage(c *gin.Context) {
	var storage model.Storage
	if err := c.ShouldBindJSON(&storage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&storage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, storage)
}

// GetStorage 获取存储详情
func GetStorage(c *gin.Context) {
	id := c.Param("id")
	var storage model.Storage
	if err := database.DB.First(&storage, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage not found"})
		return
	}
	c.JSON(http.StatusOK, storage)
}

// UpdateStorage 更新存储
func UpdateStorage(c *gin.Context) {
	id := c.Param("id")
	var storage model.Storage
	if err := database.DB.First(&storage, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage not found"})
		return
	}

	if err := c.ShouldBindJSON(&storage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Save(&storage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, storage)
}

// DeleteStorage 删除存储
func DeleteStorage(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&model.Storage{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Storage deleted successfully"})
}

// TestStorageConnection 测试存储连接
func TestStorageConnection(c *gin.Context) {
	id := c.Param("id")
	var storageModel model.Storage
	if err := database.DB.First(&storageModel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage not found"})
		return
	}

	// 测试连接
	storageSvc := service.NewStorageService()
	if err := storageSvc.TestConnection(&storageModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Storage connection test successful"})
}

// GetStorageDiskSpace 获取存储磁盘空间信息
func GetStorageDiskSpace(c *gin.Context) {
	id := c.Param("id")
	var storageModel model.Storage
	if err := database.DB.First(&storageModel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage not found"})
		return
	}

	// 只有本地存储和SSH存储支持获取磁盘空间
	if storageModel.Type != "local" && storageModel.Type != "ssh" {
		c.JSON(http.StatusOK, gin.H{
			"total":      0,
			"used":       0,
			"free":       0,
			"percentage": 0,
		})
		return
	}

	// 解析存储配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(storageModel.Config), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse storage config"})
		return
	}

	// 创建存储实例
	storageInstance, err := storage.NewStorage(storageModel.Type, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage instance"})
		return
	}

	// 获取磁盘空间
	ctx := context.Background()
	var total, used, free uint64
	var diskErr error

	if localStorage, ok := storageInstance.(*storage.LocalStorage); ok {
		total, used, free, diskErr = localStorage.GetDiskSpace(ctx)
	} else if sshStorage, ok := storageInstance.(*storage.SSHStorage); ok {
		total, used, free, diskErr = sshStorage.GetDiskSpace(ctx)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"total":      0,
			"used":       0,
			"free":       0,
			"percentage": 0,
		})
		return
	}

	if diskErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get disk space: %v", diskErr)})
		return
	}

	// 计算使用百分比
	percentage := float64(0)
	if total > 0 {
		percentage = float64(used) / float64(total) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total":      total,
		"used":       used,
		"free":       free,
		"percentage": percentage,
	})
}

// GetNotifications 获取通知列表
func GetNotifications(c *gin.Context) {
	var notifications []model.Notification
	if err := database.DB.Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}

// CreateNotification 创建通知
func CreateNotification(c *gin.Context) {
	var notification model.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetNotification 获取通知详情
func GetNotification(c *gin.Context) {
	id := c.Param("id")
	var notification model.Notification
	if err := database.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}
	c.JSON(http.StatusOK, notification)
}

// UpdateNotification 更新通知
func UpdateNotification(c *gin.Context) {
	id := c.Param("id")
	var notification model.Notification
	if err := database.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// DeleteNotification 删除通知
func DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&model.Notification{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// TestNotification 测试通知
func TestNotification(c *gin.Context) {
	id := c.Param("id")
	var notifModel model.Notification
	if err := database.DB.First(&notifModel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// 测试通知
	notifSvc := service.NewNotificationService()
	if err := notifSvc.TestNotification(&notifModel); err != nil {
		// 记录详细错误日志到日志文件
		logger.Error("Test notification failed for ID %s (type: %s): %v", id, notifModel.Type, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	logger.Info("Test notification successful for ID %s (type: %s)", id, notifModel.Type)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Notification test successful"})
}

// GetLogs 获取日志列表
func GetLogs(c *gin.Context) {
	var logs []model.BackupLog

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	// 筛选参数
	status := c.Query("status")
	taskID := c.Query("task_id")
	backupType := c.Query("backup_type")
	taskName := c.Query("task_name")
	hostName := c.Query("host_name")
	storageType := c.Query("storage_type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	query := database.DB.Model(&model.BackupLog{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}
	if backupType != "" {
		query = query.Where("backup_type = ?", backupType)
	}
	if taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if storageType != "" {
		query = query.Where("storage_type = ?", storageType)
	}
	if startTime != "" {
		query = query.Where("start_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("start_time <= ?", endTime)
	}

	var total int64
	query.Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetLog 获取日志详情
func GetLog(c *gin.Context) {
	id := c.Param("id")
	var log model.BackupLog
	if err := database.DB.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}
	c.JSON(http.StatusOK, log)
}

// DeleteLog 删除日志
func DeleteLog(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&model.BackupLog{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Log deleted successfully"})
}

// DeleteBackup 删除备份文件和日志
func DeleteBackup(c *gin.Context) {
	id := c.Param("id")

	// 查询备份日志
	var log model.BackupLog
	if err := database.DB.Preload("Task").First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup log not found"})
		return
	}

	// 删除备份文件
	if log.Status == "success" && log.FilePath != "" {
		// 加载任务信息以获取存储配置
		var task model.Task
		if err := database.DB.Preload("Storage").First(&task, log.TaskID).Error; err == nil {
			// 删除存储中的文件
			backupSvc := service.NewBackupService()
			if err := backupSvc.DeleteBackupFile(&task, log.FilePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete backup file: %v", err)})
				return
			}
		}
	}

	// 删除日志记录
	if err := database.DB.Delete(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Backup deleted successfully"})
}

// DownloadBackup 下载备份文件
func DownloadBackup(c *gin.Context) {
	id := c.Param("id")

	// 查询备份日志
	var log model.BackupLog
	if err := database.DB.Preload("Task").First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup log not found"})
		return
	}

	if log.Status != "success" || log.FilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Backup file not available"})
		return
	}

	// 加载任务和存储信息
	var task model.Task
	if err := database.DB.Preload("Storage").First(&task, log.TaskID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load task"})
		return
	}

	// 解析存储配置
	var storageConfig map[string]interface{}
	if err := json.Unmarshal([]byte(task.Storage.Config), &storageConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse storage config"})
		return
	}

	// 创建存储实例
	storageInstance, err := storage.NewStorage(task.Storage.Type, storageConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage instance"})
		return
	}

	// 下载文件到临时目录
	tmpFile := filepath.Join(os.TempDir(), filepath.Base(log.FilePath))
	ctx := context.Background()

	if err := storageInstance.Download(ctx, log.FilePath, tmpFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download file: %v", err)})
		return
	}
	defer os.Remove(tmpFile)

	// 发送文件
	c.FileAttachment(tmpFile, filepath.Base(log.FilePath))
}

// GetUsers 获取用户列表
func GetUsers(c *gin.Context) {
	var users []model.User
	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser 获取用户详情
func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUser 更新用户
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&model.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
