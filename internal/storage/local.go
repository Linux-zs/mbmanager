package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage 本地存储
type LocalStorage struct {
	basePath string
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	BasePath string `json:"base_path"`
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(config map[string]interface{}) (*LocalStorage, error) {
	basePath, ok := config["base_path"].(string)
	if !ok || basePath == "" {
		basePath = "./data/backups"
	}

	// 确保目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, localPath string, remotePath string) error {
	dstPath := filepath.Join(s.basePath, remotePath)

	// 创建目标目录
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 复制文件
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func (s *LocalStorage) Download(ctx context.Context, remotePath string, localPath string) error {
	srcPath := filepath.Join(s.basePath, remotePath)

	// 创建目标目录
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 复制文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func (s *LocalStorage) Delete(ctx context.Context, remotePath string) error {
	filePath := filepath.Join(s.basePath, remotePath)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *LocalStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	searchPath := filepath.Join(s.basePath, prefix)
	var files []FileInfo

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath := strings.TrimPrefix(path, s.basePath)
			relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

			files = append(files, FileInfo{
				Name:         info.Name(),
				Path:         relPath,
				Size:         info.Size(),
				ModifiedTime: info.ModTime(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

func (s *LocalStorage) Exists(ctx context.Context, remotePath string) (bool, error) {
	filePath := filepath.Join(s.basePath, remotePath)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) GetFileInfo(ctx context.Context, remotePath string) (*FileInfo, error) {
	filePath := filepath.Join(s.basePath, remotePath)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Name:         info.Name(),
		Path:         remotePath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
	}, nil
}

func (s *LocalStorage) TestConnection(ctx context.Context) error {
	// 检查目录是否存在且可写
	testFile := filepath.Join(s.basePath, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}
	os.Remove(testFile)
	return nil
}
