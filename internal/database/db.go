package database

import (
	"mbmanager/internal/model"
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(dbPath string) error {
	var err error

	// 打开数据库连接
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移数据库表
	err = DB.AutoMigrate(
		&model.Host{},
		&model.Storage{},
		&model.Notification{},
		&model.Task{},
		&model.BackupLog{},
		&model.User{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化默认数据
	if err := initDefaultData(); err != nil {
		return fmt.Errorf("failed to init default data: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	// 创建默认管理员用户
	var userCount int64
	DB.Model(&model.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		defaultUser := &model.User{
			Username: "admin",
			Password: string(hashedPassword),
			Email:    "admin@example.com",
			Role:     "admin",
			Status:   1,
		}

		if err := DB.Create(defaultUser).Error; err != nil {
			return err
		}
		log.Println("Default admin user created (username: admin, password: admin123)")
	}

	// 创建默认本地存储
	var storageCount int64
	DB.Model(&model.Storage{}).Count(&storageCount)
	if storageCount == 0 {
		defaultStorage := &model.Storage{
			Name:      "本地存储",
			Type:      "local",
			Config:    `{"base_path":"/data/backups"}`,
			IsDefault: 1,
			Status:    1,
		}

		if err := DB.Create(defaultStorage).Error; err != nil {
			return err
		}
		log.Println("Default local storage created")
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
