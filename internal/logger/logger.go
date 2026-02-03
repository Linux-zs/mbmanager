package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
	logFile     *os.File
)

// InitLogger 初始化日志系统
func InitLogger(logDir string) error {
	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 创建日志文件（按日期命名）
	logFileName := fmt.Sprintf("mbmanager_%s.log", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(logDir, logFileName)

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// 创建多写入器（同时输出到文件和控制台）
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// 初始化不同级别的日志记录器
	InfoLogger = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.Printf("Logger initialized, log file: %s", logFilePath)

	return nil
}

// CloseLogger 关闭日志文件
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debug 记录调试日志
func Debug(format string, v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
