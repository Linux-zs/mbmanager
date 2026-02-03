package storage

import (
	"fmt"
)

// NASStorage NAS存储（实际上是挂载的本地路径）
type NASStorage struct {
	*LocalStorage
}

// NASConfig NAS配置
type NASConfig struct {
	MountPath string `json:"mount_path"` // NFS挂载路径
}

// NewNASStorage 创建NAS存储实例
func NewNASStorage(config map[string]interface{}) (*NASStorage, error) {
	mountPath, _ := config["mount_path"].(string)
	if mountPath == "" {
		return nil, fmt.Errorf("mount_path is required")
	}

	// 使用LocalStorage实现，只是路径不同
	localConfig := map[string]interface{}{
		"base_path": mountPath,
	}

	localStorage, err := NewLocalStorage(localConfig)
	if err != nil {
		return nil, err
	}

	return &NASStorage{
		LocalStorage: localStorage,
	}, nil
}
