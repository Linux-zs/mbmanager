package model

import (
	"time"
)

// Host MySQL数据源
type Host struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Host         string    `gorm:"size:255;not null" json:"host"`
	Port         int       `gorm:"not null;default:3306" json:"port"`
	Username     string    `gorm:"size:100;not null" json:"username"`
	Password     string    `gorm:"size:255;not null" json:"password"` // 加密存储
	Group        string    `gorm:"size:100;index" json:"group"`       // 主机分组
	Description  string    `gorm:"type:text" json:"description"`
	MySQLVersion string    `gorm:"size:50" json:"mysql_version"` // MySQL版本
	Status       int       `gorm:"default:1" json:"status"`      // 1:启用 0:禁用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Host) TableName() string {
	return "hosts"
}
