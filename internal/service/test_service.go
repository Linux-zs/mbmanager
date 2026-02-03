package service

import (
	"context"
	"database/sql"
	"mbmanager/internal/model"
	"mbmanager/internal/notification"
	"mbmanager/internal/storage"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// HostService 主机服务
type HostService struct{}

// NewHostService 创建主机服务实例
func NewHostService() *HostService {
	return &HostService{}
}

// TestConnection 测试MySQL连接
func (s *HostService) TestConnection(host *model.Host) error {
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		host.Username,
		host.Password,
		host.Host,
		host.Port,
	)

	// 尝试连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 获取MySQL版本
	var version string
	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err == nil {
		host.MySQLVersion = version
	}

	return nil
}

// GetDatabases 获取数据库列表
func (s *HostService) GetDatabases(host *model.Host) ([]string, error) {
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		host.Username,
		host.Password,
		host.Host,
		host.Port,
	)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	// 查询数据库列表
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		// 过滤系统数据库
		if dbName != "information_schema" && dbName != "performance_schema" && dbName != "mysql" && dbName != "sys" {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// StorageService 存储服务
type StorageService struct{}

// NewStorageService 创建存储服务实例
func NewStorageService() *StorageService {
	return &StorageService{}
}

// TestConnection 测试存储连接
func (s *StorageService) TestConnection(storageModel *model.Storage) error {
	// 解析存储配置
	var storageConfig map[string]interface{}
	if err := json.Unmarshal([]byte(storageModel.Config), &storageConfig); err != nil {
		return fmt.Errorf("failed to parse storage config: %w", err)
	}

	// 创建存储实例
	storageInstance, err := storage.NewStorage(storageModel.Type, storageConfig)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// 测试连接
	ctx := context.Background()
	if err := storageInstance.TestConnection(ctx); err != nil {
		return fmt.Errorf("storage connection test failed: %w", err)
	}

	return nil
}

// NotificationService 通知服务
type NotificationService struct{}

// NewNotificationService 创建通知服务实例
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// TestNotification 测试通知
func (s *NotificationService) TestNotification(notifModel *model.Notification) error {
	// 解析通知配置
	var notifConfig map[string]interface{}
	if err := json.Unmarshal([]byte(notifModel.Config), &notifConfig); err != nil {
		return fmt.Errorf("failed to parse notification config: %w", err)
	}

	// 创建通知实例
	notifier, err := notification.NewNotifier(notifModel.Type, notifConfig)
	if err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}

	// 测试通知
	ctx := context.Background()
	if err := notifier.Test(ctx); err != nil {
		return fmt.Errorf("notification test failed: %w", err)
	}

	return nil
}
