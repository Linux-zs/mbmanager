package model

import (
	"time"
)

// Task 备份任务
type Task struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Name             string     `gorm:"uniqueIndex;size:100;not null" json:"name"`
	HostID           uint       `gorm:"not null;index" json:"host_id"`
	Host             *Host      `gorm:"foreignKey:HostID" json:"host,omitempty"`
	Databases        string     `gorm:"type:text" json:"databases"` // JSON数组，要备份的数据库列表
	BackupType       string     `gorm:"size:20;not null" json:"backup_type"` // mysqldump, mydumper, xtrabackup
	ScheduleType     string     `gorm:"size:20;not null" json:"schedule_type"` // once, daily, weekly, monthly, cron
	ScheduleConfig   string     `gorm:"type:text;not null" json:"schedule_config"` // JSON格式存储调度配置
	StorageID        uint       `gorm:"not null" json:"storage_id"`
	Storage          *Storage   `gorm:"foreignKey:StorageID" json:"storage,omitempty"`
	RetentionDays    int        `gorm:"default:7" json:"retention_days"` // 保留天数
	NotificationIDs  string     `gorm:"type:text" json:"notification_ids"` // JSON数组
	NotifyOnSuccess  int        `gorm:"default:0" json:"notify_on_success"`
	NotifyOnFailure  int        `gorm:"default:1" json:"notify_on_failure"`
	BackupOptions    string     `gorm:"type:text" json:"backup_options"` // JSON格式存储备份选项
	CompressionType  string     `gorm:"size:20;default:'gzip'" json:"compression_type"` // none, gzip, zip
	Status           int        `gorm:"default:1;index" json:"status"` // 1:启用 0:禁用
	LastRunAt        *time.Time `json:"last_run_at"`
	NextRunAt        *time.Time `gorm:"index" json:"next_run_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (Task) TableName() string {
	return "tasks"
}
