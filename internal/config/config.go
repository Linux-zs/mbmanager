package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Backup   BackupConfig
}

type ServerConfig struct {
	Port string
	Mode string // debug, release
}

type DatabaseConfig struct {
	Path string
}

type BackupConfig struct {
	BasePath string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/mbmanager.db"),
		},
		Backup: BackupConfig{
			BasePath: getEnv("BACKUP_PATH", "./data/backups"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
