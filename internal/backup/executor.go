package backup

import (
	"context"
	"time"
)

// Executor 备份执行器接口
type Executor interface {
	// Execute 执行备份
	Execute(ctx context.Context, params *BackupParams) (*BackupResult, error)
	// Type 获取执行器类型
	Type() string
	// Validate 验证参数
	Validate(params *BackupParams) error
}

// BackupParams 备份参数
type BackupParams struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Databases       []string               // 空表示全部数据库
	OutputPath      string                 // 输出路径
	Options         map[string]interface{} // 额外选项
	CompressionType string                 // none, gzip, zip
	SSHConfig       *SSHConfig             // xtrabackup需要
}

// SSHConfig SSH配置
type SSHConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	XtrabackupPath string `json:"xtrabackup_path"` // xtrabackup工具路径
}

// BackupResult 备份结果
type BackupResult struct {
	FilePath     string
	FileSize     int64
	Duration     time.Duration
	Databases    []string
	Command      string // 完整的备份命令
	BackupTime   int    // 备份耗时（秒）
	TransferTime int    // 传输耗时（秒）
	Error        error
}

// NewExecutor 创建备份执行器
func NewExecutor(backupType string) Executor {
	switch backupType {
	case "mysqldump":
		return &MysqldumpExecutor{}
	case "mydumper":
		return &MydumperExecutor{}
	case "xtrabackup":
		return &XtrabackupExecutor{}
	default:
		return &MysqldumpExecutor{} // 默认使用mysqldump
	}
}
