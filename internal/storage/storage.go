package storage

import (
	"context"
	"fmt"
	"time"
)

// Storage 存储接口
type Storage interface {
	// Upload 上传文件
	Upload(ctx context.Context, localPath string, remotePath string) error
	// Download 下载文件
	Download(ctx context.Context, remotePath string, localPath string) error
	// Delete 删除文件
	Delete(ctx context.Context, remotePath string) error
	// List 列出文件
	List(ctx context.Context, prefix string) ([]FileInfo, error)
	// Exists 检查文件是否存在
	Exists(ctx context.Context, remotePath string) (bool, error)
	// GetFileInfo 获取文件信息
	GetFileInfo(ctx context.Context, remotePath string) (*FileInfo, error)
	// TestConnection 测试连接
	TestConnection(ctx context.Context) error
}

// FileInfo 文件信息
type FileInfo struct {
	Name         string
	Path         string
	Size         int64
	ModifiedTime time.Time
}

// Config 存储配置
type Config struct {
	Type   string                 `json:"type"` // local, s3, oss, nas, ssh
	Params map[string]interface{} `json:"params"` // 具体配置参数
}

// NewStorage 创建存储实例
func NewStorage(storageType string, config map[string]interface{}) (Storage, error) {
	switch storageType {
	case "local":
		return NewLocalStorage(config)
	case "s3":
		return NewS3Storage(config)
	case "oss":
		return NewOSSStorage(config)
	case "nas":
		return NewNASStorage(config)
	case "ssh":
		return NewSSHStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
