package backup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// XtrabackupExecutor xtrabackup备份执行器（通过SSH远程执行）
type XtrabackupExecutor struct{}

func (e *XtrabackupExecutor) Type() string {
	return "xtrabackup"
}

func (e *XtrabackupExecutor) Validate(params *BackupParams) error {
	if params.Host == "" {
		return fmt.Errorf("host is required")
	}
	if params.Username == "" {
		return fmt.Errorf("username is required")
	}
	if params.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if params.SSHConfig == nil {
		return fmt.Errorf("SSH config is required for xtrabackup")
	}
	if params.SSHConfig.Host == "" {
		return fmt.Errorf("SSH host is required")
	}
	if params.SSHConfig.Username == "" {
		return fmt.Errorf("SSH username is required")
	}
	return nil
}

func (e *XtrabackupExecutor) Execute(ctx context.Context, params *BackupParams) (*BackupResult, error) {
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

	// 建立SSH连接
	client, err := e.connectSSH(params.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect SSH: %w", err)
	}
	defer client.Close()

	// 在远程服务器创建临时目录
	remoteTmpDir := fmt.Sprintf("/tmp/xtrabackup_%s", timestamp)
	if err := e.executeSSHCommand(client, fmt.Sprintf("mkdir -p %s", remoteTmpDir)); err != nil {
		return nil, fmt.Errorf("failed to create remote directory: %w", err)
	}

	// 构建xtrabackup命令
	xtrabackupPath := params.SSHConfig.XtrabackupPath
	if xtrabackupPath == "" {
		xtrabackupPath = "xtrabackup" // 默认使用PATH中的xtrabackup
	}

	cmd := fmt.Sprintf("%s --backup --host=%s --port=%d --user=%s --password='%s' --target-dir=%s",
		xtrabackupPath, params.Host, params.Port, params.Username, params.Password, remoteTmpDir)

	// 执行备份命令
	if err := e.executeSSHCommand(client, cmd); err != nil {
		e.executeSSHCommand(client, fmt.Sprintf("rm -rf %s", remoteTmpDir))
		return nil, fmt.Errorf("xtrabackup backup failed: %w", err)
	}

	// 根据压缩类型打包备份文件
	var backupFile string
	var tarCmd string
	switch params.CompressionType {
	case "none":
		// 不压缩，只打包tar
		backupFile = fmt.Sprintf("%s.tar", remoteTmpDir)
		tarCmd = fmt.Sprintf("tar -cf %s -C %s .", backupFile, remoteTmpDir)
	case "zip":
		// ZIP压缩
		backupFile = fmt.Sprintf("%s.zip", remoteTmpDir)
		tarCmd = fmt.Sprintf("cd %s && zip -r %s .", remoteTmpDir, backupFile)
	default: // gzip
		// GZIP压缩（tar.gz）
		backupFile = fmt.Sprintf("%s.tar.gz", remoteTmpDir)
		tarCmd = fmt.Sprintf("tar -czf %s -C %s .", backupFile, remoteTmpDir)
	}

	if err := e.executeSSHCommand(client, tarCmd); err != nil {
		e.executeSSHCommand(client, fmt.Sprintf("rm -rf %s", remoteTmpDir))
		return nil, fmt.Errorf("failed to compress backup: %w", err)
	}

	// 下载备份文件到本地
	localFile := filepath.Join(outputDir, filepath.Base(backupFile))
	if err := e.downloadFile(client, backupFile, localFile); err != nil {
		e.executeSSHCommand(client, fmt.Sprintf("rm -rf %s %s", remoteTmpDir, backupFile))
		return nil, fmt.Errorf("failed to download backup: %w", err)
	}

	// 清理远程文件
	e.executeSSHCommand(client, fmt.Sprintf("rm -rf %s %s", remoteTmpDir, backupFile))

	// 获取文件大小
	fileInfo, err := os.Stat(localFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &BackupResult{
		FilePath:  localFile,
		FileSize:  fileInfo.Size(),
		Duration:  time.Since(startTime),
		Databases: params.Databases,
	}, nil
}

// connectSSH 建立SSH连接
func (e *XtrabackupExecutor) connectSSH(config *SSHConfig) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// 密码认证
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	// 私钥认证
	if config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
		Timeout:         30 * time.Second,
	}

	port := config.Port
	if port == 0 {
		port = 22
	}

	address := fmt.Sprintf("%s:%d", config.Host, port)
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	return client, nil
}

// executeSSHCommand 执行SSH命令
func (e *XtrabackupExecutor) executeSSHCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

// downloadFile 通过SCP下载文件
func (e *XtrabackupExecutor) downloadFile(client *ssh.Client, remotePath, localPath string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// 使用cat命令读取远程文件
	output, err := session.Output(fmt.Sprintf("cat %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to read remote file: %w", err)
	}

	// 写入本地文件
	if err := os.WriteFile(localPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}
