package service

import (
	"context"
	"mbmanager/internal/backup"
	"mbmanager/internal/database"
	"mbmanager/internal/model"
	"mbmanager/internal/notification"
	"mbmanager/internal/storage"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// BackupService 备份服务
type BackupService struct{}

// NewBackupService 创建备份服务实例
func NewBackupService() *BackupService {
	return &BackupService{}
}

// ExecuteBackup 执行备份任务
func (s *BackupService) ExecuteBackup(ctx context.Context, task *model.Task) error {
	log.Printf("Starting backup task: %s (ID: %d)", task.Name, task.ID)

	// 创建备份日志
	backupLog := &model.BackupLog{
		TaskID:     task.ID,
		TaskName:   task.Name,
		BackupType: task.BackupType,
		Status:     "running",
		StartTime:  time.Now(),
	}

	// 加载主机信息
	var host model.Host
	if err := database.DB.First(&host, task.HostID).Error; err != nil {
		backupLog.Status = "failed"
		backupLog.ErrorMessage = fmt.Sprintf("Failed to load host: %v", err)
		s.saveLog(backupLog)
		return err
	}
	backupLog.HostName = host.Name

	// 解析数据库列表
	var databases []string
	if task.Databases != "" {
		if err := json.Unmarshal([]byte(task.Databases), &databases); err != nil {
			log.Printf("Failed to parse databases: %v", err)
		}
	}
	backupLog.Databases = task.Databases

	// 保存初始日志
	if err := database.DB.Create(backupLog).Error; err != nil {
		log.Printf("Failed to create backup log: %v", err)
	}

	// 执行备份
	result, err := s.performBackup(ctx, task, &host, databases)

	// 更新日志
	endTime := time.Now()
	backupLog.EndTime = &endTime
	backupLog.Duration = int(endTime.Sub(backupLog.StartTime).Seconds())

	if err != nil {
		backupLog.Status = "failed"
		backupLog.ErrorMessage = err.Error()
		database.DB.Save(backupLog)

		// 发送失败通知
		if task.NotifyOnFailure == 1 {
			s.sendNotification(task, backupLog)
		}

		return err
	}

	// 备份成功
	backupLog.Status = "success"
	backupLog.FilePath = result.FilePath
	backupLog.FileSize = result.FileSize
	backupLog.Command = result.Command
	backupLog.BackupTime = result.BackupTime
	backupLog.TransferTime = result.TransferTime

	// 加载存储信息
	var storageModel model.Storage
	if err := database.DB.First(&storageModel, task.StorageID).Error; err == nil {
		backupLog.StorageType = storageModel.Type
		backupLog.StorageName = storageModel.Name
	}

	database.DB.Save(backupLog)

	// 更新任务状态
	now := time.Now()
	task.LastRunAt = &now
	database.DB.Save(task)

	// 清理过期备份
	s.cleanupExpiredBackups(task)

	// 发送成功通知
	if task.NotifyOnSuccess == 1 {
		s.sendNotification(task, backupLog)
	}

	log.Printf("Backup task completed: %s (ID: %d)", task.Name, task.ID)
	return nil
}

// performBackup 执行备份
func (s *BackupService) performBackup(ctx context.Context, task *model.Task, host *model.Host, databases []string) (*backup.BackupResult, error) {
	// 创建临时目录
	tmpDir := filepath.Join("./data/tmp", fmt.Sprintf("backup_%d_%d", task.ID, time.Now().Unix()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir) // 清理临时目录

	// 解析备份选项
	backupOptions := make(map[string]interface{})
	if task.BackupOptions != "" {
		// 对于xtrabackup，尝试解析为JSON（包含SSH配置）
		if task.BackupType == "xtrabackup" {
			if err := json.Unmarshal([]byte(task.BackupOptions), &backupOptions); err != nil {
				// 如果解析失败，将其作为命令行参数
				backupOptions["extra_args"] = task.BackupOptions
			}
		} else {
			// 对于mysqldump和mydumper，将命令行参数字符串保存到extra_args
			backupOptions["extra_args"] = task.BackupOptions
		}
	}

	// 准备备份参数
	params := &backup.BackupParams{
		Host:            host.Host,
		Port:            host.Port,
		Username:        host.Username,
		Password:        host.Password,
		Databases:       databases,
		OutputPath:      tmpDir,
		Options:         backupOptions,
		CompressionType: task.CompressionType,
	}

	// 如果是xtrabackup，需要SSH配置
	if task.BackupType == "xtrabackup" {
		if sshConfig, ok := backupOptions["ssh_config"].(map[string]interface{}); ok {
			params.SSHConfig = &backup.SSHConfig{
				Host:           getStringValue(sshConfig, "host"),
				Port:           getIntValue(sshConfig, "port"),
				Username:       getStringValue(sshConfig, "username"),
				Password:       getStringValue(sshConfig, "password"),
				PrivateKey:     getStringValue(sshConfig, "private_key"),
				XtrabackupPath: getStringValue(sshConfig, "xtrabackup_path"),
			}
		}
	}

	// 创建备份执行器
	executor := backup.NewExecutor(task.BackupType)

	// 执行备份并记录时间
	backupStartTime := time.Now()
	result, err := executor.Execute(ctx, params)
	backupDuration := int(time.Since(backupStartTime).Seconds())

	if err != nil {
		return nil, fmt.Errorf("backup execution failed: %w", err)
	}

	// 记录备份耗时
	result.BackupTime = backupDuration

	// 上传到存储并记录时间
	transferStartTime := time.Now()
	remotePath, err := s.uploadToStorage(task, result.FilePath, host.Name)
	transferDuration := int(time.Since(transferStartTime).Seconds())

	if err != nil {
		return nil, fmt.Errorf("failed to upload backup: %w", err)
	}

	// 记录传输耗时
	result.TransferTime = transferDuration

	// 更新结果中的文件路径为存储路径
	result.FilePath = remotePath

	return result, nil
}

// uploadToStorage 上传备份文件到存储，返回远程路径
func (s *BackupService) uploadToStorage(task *model.Task, localPath string, hostName string) (string, error) {
	// 加载存储配置
	var storageModel model.Storage
	if err := database.DB.First(&storageModel, task.StorageID).Error; err != nil {
		return "", fmt.Errorf("failed to load storage: %w", err)
	}

	// 解析存储配置
	var storageConfig map[string]interface{}
	if err := json.Unmarshal([]byte(storageModel.Config), &storageConfig); err != nil {
		return "", fmt.Errorf("failed to parse storage config: %w", err)
	}

	// 创建存储实例
	storageInstance, err := storage.NewStorage(storageModel.Type, storageConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create storage: %w", err)
	}

	// 生成远程路径（使用主机名而不是task_ID）
	remotePath := filepath.Join(
		hostName,
		filepath.Base(localPath),
	)

	// 上传文件
	ctx := context.Background()
	if err := storageInstance.Upload(ctx, localPath, remotePath); err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return remotePath, nil
}

// cleanupExpiredBackups 清理过期备份
func (s *BackupService) cleanupExpiredBackups(task *model.Task) {
	if task.RetentionDays <= 0 {
		return
	}

	// 计算过期时间
	expireTime := time.Now().AddDate(0, 0, -task.RetentionDays)

	// 查询过期的备份日志
	var expiredLogs []model.BackupLog
	database.DB.Where("task_id = ? AND status = ? AND start_time < ?",
		task.ID, "success", expireTime).Find(&expiredLogs)

	// 删除过期备份文件和日志
	for _, expiredLog := range expiredLogs {
		// 删除备份文件（从存储中）
		if err := s.DeleteBackupFile(task, expiredLog.FilePath); err != nil {
			log.Printf("Failed to delete backup file: %v", err)
		}

		// 删除日志记录
		database.DB.Delete(&expiredLog)
	}

	if len(expiredLogs) > 0 {
		log.Printf("Cleaned up %d expired backups for task %s", len(expiredLogs), task.Name)
	}
}

// DeleteBackupFile 删除备份文件
func (s *BackupService) DeleteBackupFile(task *model.Task, filePath string) error {
	// 加载主机信息
	var host model.Host
	if err := database.DB.First(&host, task.HostID).Error; err != nil {
		return err
	}

	// 加载存储配置
	var storageModel model.Storage
	if err := database.DB.First(&storageModel, task.StorageID).Error; err != nil {
		return err
	}

	// 解析存储配置
	var storageConfig map[string]interface{}
	if err := json.Unmarshal([]byte(storageModel.Config), &storageConfig); err != nil {
		return err
	}

	// 创建存储实例
	storageInstance, err := storage.NewStorage(storageModel.Type, storageConfig)
	if err != nil {
		return err
	}

	// 删除文件（使用主机名而不是task_ID）
	ctx := context.Background()
	remotePath := filepath.Join(
		host.Name,
		filepath.Base(filePath),
	)

	return storageInstance.Delete(ctx, remotePath)
}

// sendNotification 发送通知
func (s *BackupService) sendNotification(task *model.Task, backupLog *model.BackupLog) {
	if task.NotificationIDs == "" {
		return
	}

	// 解析通知ID列表
	var notificationIDs []int
	if err := json.Unmarshal([]byte(task.NotificationIDs), &notificationIDs); err != nil {
		log.Printf("Failed to parse notification IDs: %v", err)
		return
	}

	// 构建通知内容
	var databases []string
	if backupLog.Databases != "" {
		json.Unmarshal([]byte(backupLog.Databases), &databases)
	}

	backupNotif := &notification.BackupNotification{
		TaskName:     backupLog.TaskName,
		HostName:     backupLog.HostName,
		Databases:    databases,
		BackupType:   backupLog.BackupType,
		Status:       backupLog.Status,
		StartTime:    backupLog.StartTime,
		EndTime:      *backupLog.EndTime,
		Duration:     time.Duration(backupLog.Duration) * time.Second,
		FileSize:     backupLog.FileSize,
		ErrorMessage: backupLog.ErrorMessage,
	}

	message := backupNotif.ToMessage()

	// 发送到所有配置的通知渠道
	for _, notifID := range notificationIDs {
		var notifModel model.Notification
		if err := database.DB.First(&notifModel, notifID).Error; err != nil {
			log.Printf("Failed to load notification %d: %v", notifID, err)
			continue
		}

		// 解析通知配置
		var notifConfig map[string]interface{}
		if err := json.Unmarshal([]byte(notifModel.Config), &notifConfig); err != nil {
			log.Printf("Failed to parse notification config: %v", err)
			continue
		}

		// 创建通知实例
		notifier, err := notification.NewNotifier(notifModel.Type, notifConfig)
		if err != nil {
			log.Printf("Failed to create notifier: %v", err)
			continue
		}

		// 发送通知
		ctx := context.Background()
		if err := notifier.Send(ctx, message); err != nil {
			log.Printf("Failed to send notification: %v", err)
		} else {
			log.Printf("Notification sent successfully to %s", notifModel.Name)
		}
	}
}

// saveLog 保存日志
func (s *BackupService) saveLog(backupLog *model.BackupLog) {
	if backupLog.EndTime == nil {
		now := time.Now()
		backupLog.EndTime = &now
		backupLog.Duration = int(now.Sub(backupLog.StartTime).Seconds())
	}
	database.DB.Save(backupLog)
}

// 辅助函数
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getIntValue(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}
