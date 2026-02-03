package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHStorage SSH远程存储
type SSHStorage struct {
	host       string
	port       int
	username   string
	password   string
	privateKey string
	basePath   string
}

// SSHConfig SSH存储配置
type SSHConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	BasePath   string `json:"base_path"`
}

// NewSSHStorage 创建SSH存储实例
func NewSSHStorage(config map[string]interface{}) (*SSHStorage, error) {
	host, _ := config["host"].(string)
	if host == "" {
		return nil, fmt.Errorf("SSH host is required")
	}

	username, _ := config["username"].(string)
	if username == "" {
		return nil, fmt.Errorf("SSH username is required")
	}

	port := 22
	if p, ok := config["port"].(float64); ok {
		port = int(p)
	}

	password, _ := config["password"].(string)
	privateKey, _ := config["private_key"].(string)

	if password == "" && privateKey == "" {
		return nil, fmt.Errorf("SSH password or private key is required")
	}

	basePath, _ := config["base_path"].(string)
	if basePath == "" {
		basePath = "/data/backups"
	}

	return &SSHStorage{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		privateKey: privateKey,
		basePath:   basePath,
	}, nil
}

// connectSSH 建立SSH连接
func (s *SSHStorage) connectSSH() (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// 密码认证
	if s.password != "" {
		authMethods = append(authMethods, ssh.Password(s.password))
	}

	// 私钥认证
	if s.privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(s.privateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	config := &ssh.ClientConfig{
		User:            s.username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", s.host, s.port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect SSH: %w", err)
	}

	return client, nil
}

// executeCommand 执行SSH命令
func (s *SSHStorage) executeCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (s *SSHStorage) Upload(ctx context.Context, localPath string, remotePath string) error {
	client, err := s.connectSSH()
	if err != nil {
		return err
	}
	defer client.Close()

	// 构建完整的远程路径
	fullRemotePath := filepath.Join(s.basePath, remotePath)

	// 创建远程目录
	remoteDir := filepath.Dir(fullRemotePath)
	if err := s.executeCommand(client, fmt.Sprintf("mkdir -p %s", remoteDir)); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// 使用SCP上传文件
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// 打开本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// 获取文件信息
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// 使用SCP协议上传
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C0644 %d %s\n", fileInfo.Size(), filepath.Base(fullRemotePath))
		io.Copy(w, localFile)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", fullRemotePath)); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *SSHStorage) Download(ctx context.Context, remotePath string, localPath string) error {
	client, err := s.connectSSH()
	if err != nil {
		return err
	}
	defer client.Close()

	// 构建完整的远程路径
	fullRemotePath := filepath.Join(s.basePath, remotePath)

	// 创建本地目录
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// 使用cat命令读取远程文件
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("cat %s", fullRemotePath))
	if err != nil {
		return fmt.Errorf("failed to read remote file: %w", err)
	}

	// 写入本地文件
	if err := os.WriteFile(localPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}

func (s *SSHStorage) Delete(ctx context.Context, remotePath string) error {
	client, err := s.connectSSH()
	if err != nil {
		return err
	}
	defer client.Close()

	fullRemotePath := filepath.Join(s.basePath, remotePath)
	if err := s.executeCommand(client, fmt.Sprintf("rm -f %s", fullRemotePath)); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *SSHStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	client, err := s.connectSSH()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	searchPath := filepath.Join(s.basePath, prefix)

	// 使用find命令列出文件
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("find %s -type f -printf '%%p\t%%s\t%%T@\n' 2>/dev/null || true", searchPath))
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var files []FileInfo

	// 解析输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}

		fullPath := parts[0]
		var size int64
		var modTimeFloat float64
		fmt.Sscanf(parts[1], "%d", &size)
		fmt.Sscanf(parts[2], "%f", &modTimeFloat)

		// 计算相对路径
		relPath := strings.TrimPrefix(fullPath, s.basePath)
		relPath = strings.TrimPrefix(relPath, "/")

		files = append(files, FileInfo{
			Name:         filepath.Base(fullPath),
			Path:         relPath,
			Size:         size,
			ModifiedTime: time.Unix(int64(modTimeFloat), 0),
		})
	}

	return files, nil
}

func (s *SSHStorage) Exists(ctx context.Context, remotePath string) (bool, error) {
	client, err := s.connectSSH()
	if err != nil {
		return false, err
	}
	defer client.Close()

	fullRemotePath := filepath.Join(s.basePath, remotePath)

	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	err = session.Run(fmt.Sprintf("test -f %s", fullRemotePath))
	return err == nil, nil
}

func (s *SSHStorage) GetFileInfo(ctx context.Context, remotePath string) (*FileInfo, error) {
	client, err := s.connectSSH()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	fullRemotePath := filepath.Join(s.basePath, remotePath)

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("stat -c '%%s %%Y' %s", fullRemotePath))
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	var size int64
	var modTime int64
	fmt.Sscanf(string(output), "%d %d", &size, &modTime)

	return &FileInfo{
		Name:         filepath.Base(remotePath),
		Path:         remotePath,
		Size:         size,
		ModifiedTime: time.Unix(modTime, 0),
	}, nil
}

func (s *SSHStorage) TestConnection(ctx context.Context) error {
	client, err := s.connectSSH()
	if err != nil {
		return err
	}
	defer client.Close()

	// 测试创建目录
	testDir := filepath.Join(s.basePath, ".test")
	if err := s.executeCommand(client, fmt.Sprintf("mkdir -p %s && rmdir %s", testDir, testDir)); err != nil {
		return fmt.Errorf("failed to test connection: %w", err)
	}

	return nil
}

// GetDiskSpace 获取SSH存储的磁盘空间信息
func (s *SSHStorage) GetDiskSpace(ctx context.Context) (total, used, free uint64, err error) {
	client, err := s.connectSSH()
	if err != nil {
		return 0, 0, 0, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// 使用df命令获取磁盘空间，输出格式：总大小 已用 可用
	// -B1 表示以字节为单位输出
	output, err := session.Output(fmt.Sprintf("df -B1 %s | tail -1 | awk '{print $2,$3,$4}'", s.basePath))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get disk space: %w", err)
	}

	// 解析输出
	_, err = fmt.Sscanf(string(output), "%d %d %d", &total, &used, &free)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse disk space output: %w", err)
	}

	return total, used, free, nil
}
