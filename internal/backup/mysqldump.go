package backup

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MysqldumpExecutor mysqldump备份执行器
type MysqldumpExecutor struct{}

func (e *MysqldumpExecutor) Type() string {
	return "mysqldump"
}

func (e *MysqldumpExecutor) Validate(params *BackupParams) error {
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

func (e *MysqldumpExecutor) Execute(ctx context.Context, params *BackupParams) (*BackupResult, error) {
	startTime := time.Now()

	if err := e.Validate(params); err != nil {
		return nil, err
	}

	// 创建输出目录
	if err := os.MkdirAll(params.OutputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// 构建命令参数
	args := []string{
		fmt.Sprintf("--host=%s", params.Host),
		fmt.Sprintf("--port=%d", params.Port),
		fmt.Sprintf("--user=%s", params.Username),
		fmt.Sprintf("--password=%s", params.Password),
		"--single-transaction",
		"--quick",
		"--lock-tables=false",
		"--routines",
		"--triggers",
		"--events",
	}

	// 添加额外选项（命令行参数字符串）
	if options, ok := params.Options["extra_args"].(string); ok && options != "" {
		// 分割命令行参数字符串
		extraArgs := strings.Fields(options)
		args = append(args, extraArgs...)
	}

	// 添加数据库
	if len(params.Databases) > 0 {
		args = append(args, "--databases")
		args = append(args, params.Databases...)
	} else {
		args = append(args, "--all-databases")
	}

	// 创建输出文件
	timestamp := time.Now().Format("20060102_150405")
	var outputFile string
	var finalFile string

	// 根据压缩类型确定文件扩展名
	switch params.CompressionType {
	case "none":
		outputFile = filepath.Join(params.OutputPath, fmt.Sprintf("backup_%s.sql", timestamp))
		finalFile = outputFile
	case "zip":
		outputFile = filepath.Join(params.OutputPath, fmt.Sprintf("backup_%s.sql", timestamp))
		finalFile = outputFile + ".zip"
	default: // gzip
		outputFile = filepath.Join(params.OutputPath, fmt.Sprintf("backup_%s.sql", timestamp))
		finalFile = outputFile + ".gz"
	}

	// 执行mysqldump命令
	cmd := exec.CommandContext(ctx, "mysqldump", args...)

	// 构建完整命令字符串（用于日志，隐藏密码）
	cmdStr := "mysqldump"
	for _, arg := range args {
		if strings.Contains(arg, "--password=") {
			cmdStr += " --password=***"
		} else {
			cmdStr += " " + arg
		}
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.Remove(outputFile)
		return nil, fmt.Errorf("mysqldump failed: %v, stderr: %s", err, stderr.String())
	}

	// 根据压缩类型处理文件
	switch params.CompressionType {
	case "none":
		// 不压缩，直接使用原始文件
	case "zip":
		// ZIP压缩
		if err := compressFileZip(outputFile, finalFile); err != nil {
			os.Remove(outputFile)
			return nil, fmt.Errorf("failed to compress file: %w", err)
		}
		os.Remove(outputFile)
	default: // gzip
		// GZIP压缩
		if err := compressFile(outputFile, finalFile); err != nil {
			os.Remove(outputFile)
			return nil, fmt.Errorf("failed to compress file: %w", err)
		}
		os.Remove(outputFile)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(finalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &BackupResult{
		FilePath:  finalFile,
		FileSize:  fileInfo.Size(),
		Duration:  time.Since(startTime),
		Databases: params.Databases,
		Command:   cmdStr,
	}, nil
}

// compressFile 压缩文件（gzip）
func compressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzipWriter := gzip.NewWriter(dstFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, srcFile)
	return err
}

// compressFileZip 压缩文件（zip）
func compressFileZip(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	zipWriter := zip.NewWriter(dstFile)
	defer zipWriter.Close()

	// 获取源文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 创建ZIP文件头
	header, err := zip.FileInfoHeader(srcInfo)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(src)
	header.Method = zip.Deflate

	// 创建ZIP文件条目
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// 复制文件内容
	_, err = io.Copy(writer, srcFile)
	return err
}
