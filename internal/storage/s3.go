package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Storage S3存储
type S3Storage struct {
	sess       *session.Session
	s3Client   *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
}

// S3Config S3配置
type S3Config struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	UseSSL          bool   `json:"use_ssl"`
}

// NewS3Storage 创建S3存储实例
func NewS3Storage(config map[string]interface{}) (*S3Storage, error) {
	endpoint, _ := config["endpoint"].(string)
	accessKeyID, _ := config["access_key_id"].(string)
	secretAccessKey, _ := config["secret_access_key"].(string)
	bucket, _ := config["bucket"].(string)
	region, _ := config["region"].(string)

	if bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	if region == "" {
		region = "us-east-1"
	}

	// 创建会话
	awsConfig := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	}

	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3Storage{
		sess:       sess,
		s3Client:   s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucket:     bucket,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, localPath string, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *S3Storage) Download(ctx context.Context, remotePath string, localPath string) error {
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = s.downloader.DownloadWithContext(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
	})

	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func (s *S3Storage) Delete(ctx context.Context, remotePath string) error {
	_, err := s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *S3Storage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	var files []FileInfo

	err := s.s3Client.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			files = append(files, FileInfo{
				Name:         *obj.Key,
				Path:         *obj.Key,
				Size:         *obj.Size,
				ModifiedTime: *obj.LastModified,
			})
		}
		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

func (s *S3Storage) Exists(ctx context.Context, remotePath string) (bool, error) {
	_, err := s.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
	})

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (s *S3Storage) GetFileInfo(ctx context.Context, remotePath string) (*FileInfo, error) {
	result, err := s.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Name:         remotePath,
		Path:         remotePath,
		Size:         *result.ContentLength,
		ModifiedTime: *result.LastModified,
	}, nil
}

func (s *S3Storage) TestConnection(ctx context.Context) error {
	// 尝试列出bucket中的对象
	_, err := s.s3Client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		MaxKeys: aws.Int64(1),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to S3: %w", err)
	}

	return nil
}
