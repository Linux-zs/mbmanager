package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage 阿里云OSS存储
type OSSStorage struct {
	client *oss.Client
	bucket *oss.Bucket
}

// OSSConfig OSS配置
type OSSConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Bucket          string `json:"bucket"`
}

// NewOSSStorage 创建OSS存储实例
func NewOSSStorage(config map[string]interface{}) (*OSSStorage, error) {
	endpoint, _ := config["endpoint"].(string)
	accessKeyID, _ := config["access_key_id"].(string)
	accessKeySecret, _ := config["access_key_secret"].(string)
	bucketName, _ := config["bucket"].(string)

	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if bucketName == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	// 创建OSS客户端
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 获取bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	return &OSSStorage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *OSSStorage) Upload(ctx context.Context, localPath string, remotePath string) error {
	err := s.bucket.PutObjectFromFile(remotePath, localPath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

func (s *OSSStorage) Download(ctx context.Context, remotePath string, localPath string) error {
	err := s.bucket.GetObjectToFile(remotePath, localPath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

func (s *OSSStorage) Delete(ctx context.Context, remotePath string) error {
	err := s.bucket.DeleteObject(remotePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *OSSStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	var files []FileInfo

	marker := ""
	for {
		lor, err := s.bucket.ListObjects(oss.Prefix(prefix), oss.Marker(marker))
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, obj := range lor.Objects {
			files = append(files, FileInfo{
				Name:         obj.Key,
				Path:         obj.Key,
				Size:         obj.Size,
				ModifiedTime: obj.LastModified,
			})
		}

		if !lor.IsTruncated {
			break
		}
		marker = lor.NextMarker
	}

	return files, nil
}

func (s *OSSStorage) Exists(ctx context.Context, remotePath string) (bool, error) {
	exists, err := s.bucket.IsObjectExist(remotePath)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return exists, nil
}

func (s *OSSStorage) GetFileInfo(ctx context.Context, remotePath string) (*FileInfo, error) {
	meta, err := s.bucket.GetObjectMeta(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// 获取文件大小
	var size int64
	if contentLength := meta.Get("Content-Length"); contentLength != "" {
		fmt.Sscanf(contentLength, "%d", &size)
	}

	// 获取修改时间
	var modTime time.Time
	if lastModified := meta.Get("Last-Modified"); lastModified != "" {
		modTime, _ = time.Parse(time.RFC1123, lastModified)
	}

	return &FileInfo{
		Name:         remotePath,
		Path:         remotePath,
		Size:         size,
		ModifiedTime: modTime,
	}, nil
}

func (s *OSSStorage) TestConnection(ctx context.Context) error {
	// 尝试列出bucket中的对象
	_, err := s.bucket.ListObjects(oss.MaxKeys(1))
	if err != nil {
		return fmt.Errorf("failed to connect to OSS: %w", err)
	}
	return nil
}
