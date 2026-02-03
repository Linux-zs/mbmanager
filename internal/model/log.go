package model

import (
	"time"
)

// BackupLog 备份日志
type BackupLog struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	TaskID       uint       `gorm:"not null;index" json:"task_id"`
	Task         *Task      `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	TaskName     string     `gorm:"size:100" json:"task_name"`
	HostName     string     `gorm:"size:100" json:"host_name"`
	Databases    string     `gorm:"type:text" json:"databases"`
	BackupType   string     `gorm:"size:20" json:"backup_type"`
	Status       string     `gorm:"size:20;not null;index" json:"status"` // running, success, failed
	StartTime    time.Time  `gorm:"not null;index" json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Duration     int        `json:"duration"`         // 总耗时（秒）
	BackupTime   int        `json:"backup_time"`      // 备份耗时（秒）
	TransferTime int        `json:"transfer_time"`    // 传输耗时（秒）
	FilePath     string     `gorm:"type:text" json:"file_path"`
	FileSize     int64      `json:"file_size"` // 字节
	StorageType  string     `gorm:"size:20" json:"storage_type"`
	StorageName  string     `gorm:"size:100" json:"storage_name"`
	Command      string     `gorm:"type:text" json:"command"` // 完整的备份命令
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (BackupLog) TableName() string {
	return "backup_logs"
}
