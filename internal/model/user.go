package model

import (
	"time"
)

// User 用户
type User struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Username    string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password    string     `gorm:"size:255;not null" json:"-"` // bcrypt加密，不返回给前端
	Email       string     `gorm:"size:100" json:"email"`
	Role        string     `gorm:"size:20;default:admin" json:"role"` // admin, viewer
	Status      int        `gorm:"default:1" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
