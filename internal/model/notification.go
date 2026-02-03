package model

import (
	"time"
)

// Notification 通知配置
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Type      string    `gorm:"size:20;not null" json:"type"` // email, dingtalk, wecom
	Config    string    `gorm:"type:text;not null" json:"config"` // JSON格式存储配置
	IsDefault int       `gorm:"default:0" json:"is_default"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
