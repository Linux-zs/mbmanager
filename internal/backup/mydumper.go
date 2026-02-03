package backup

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// MydumperExecutor mydumper备份执行器
type MydumperExecutor struct{}

func (e *MydumperExecutor) Type() string {
	return "mydumper"
}

func (e *MydumperExecutor) Validate(params *BackupParams) error {
	if params.Host == "" {
		return fmt.Errorf("host is required")
	}
	if params.Username == "" {
		return fmt.Errorf("username is required")
	}
	if params.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}
	return nil
}

func (e *MydumperExecutor) Execute(ctx context.Context, params *BackupParams) (*BackupResult, error) {
	startTime := time.Now()

	if err := e.Validate(params); err != nil {
		return nil, err
	}

	// 创建输出目录
	timestamp := time.Now().Format("20060102_150405")
	outputDir := filepath.Join(params.OutputPath, fmt.Sprintf("backup_%s", timestamp))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// 构建命令参数（不使用--compress，因为最后会打包成tar.gz）
	args := []string{
		"-h", params.Host,
		"-P", fmt.Sprintf("%d", params.Port),
		"-u", params.Username,
		"-p", params.Password,
		"-o", outputDir,
		"--threads", "4",
	}

	// 添加额外选项
	if options, ok := params.Options["threads"].(string); ok && options != "" {
		args = append(args, "--threads", options)
	}

	// 添加数据库
	if len(params.Databases) > 0 {
		for _, db := range params.Databases {
			args = append(args, "-B", db)
		}
	}

	// 执行mydumper命令
	cmd := exec.CommandContext(ctx, "mydumper", args...)

	// 构建完整命令字符串（用于日志，隐藏密码）
	cmdStr := "mydumper"
	for _, arg := range args {
		if arg == params.Password {
			cmdStr += " ***"
		} else {
			cmdStr += " " + arg
		}
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.RemoveAll(outputDir)
		return nil, fmt.Errorf("mydumper failed: %v, stderr: %s", err, stderr.String())
	}

	// 根据压缩类型处理目录
	var finalPath string
	switch params.CompressionType {
	case "none":
		// 不压缩，直接使用目录（但需要打包成tar以便传输）
		finalPath = outputDir + ".tar"
		if err := createTar(outputDir, finalPath); err != nil {
			os.RemoveAll(outputDir)
			return nil, fmt.Errorf("failed to create tar: %w", err)
		}
	case "zip":
		// ZIP压缩
		finalPath = outputDir + ".zip"
		if err := createZip(outputDir, finalPath); err != nil {
			os.RemoveAll(outputDir)
			return nil, fmt.Errorf("failed to create zip: %w", err)
		}
	default: // gzip
		// GZIP压缩（tar.gz）
		finalPath = outputDir + ".tar.gz"
		if err := createTarGz(outputDir, finalPath); err != nil {
			os.RemoveAll(outputDir)
			return nil, fmt.Errorf("failed to create tar.gz: %w", err)
		}
	}

	// 删除原始目录
	os.RemoveAll(outputDir)

	// 获取文件大小
	fileInfo, err := os.Stat(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &BackupResult{
		FilePath:  finalPath,
		FileSize:  fileInfo.Size(),
		Duration:  time.Since(startTime),
		Databases: params.Databases,
		Command:   cmdStr,
	}, nil
}

// createTarGz 将目录打包成tar.gz文件
func createTarGz(sourceDir, targetFile string) error {
	// 创建tar.gz文件
	file, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %w", err)
	}
	defer file.Close()

	// 创建gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// 创建tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 遍历目录并添加到tar
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// 设置相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		// 写入header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// createTar 将目录打包成tar文件（不压缩）
func createTar(sourceDir, targetFile string) error {
	// 创建tar文件
	file, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer file.Close()

	// 创建tar writer
	tarWriter := tar.NewWriter(file)
	defer tarWriter.Close()

	// 遍历目录并添加到tar
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// 设置相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		// 写入header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// createZip 将目录打包成zip文件
func createZip(sourceDir, targetFile string) error {
	// 创建zip文件
	file, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer file.Close()

	// 创建zip writer
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 遍历目录并添加到zip
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 设置相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		// 设置压缩方法
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		// 创建zip文件条目
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}

		return nil
	})
}
