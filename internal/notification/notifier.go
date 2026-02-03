package notification

import (
	"context"
	"fmt"
	"time"
)

// Notifier 通知接口
type Notifier interface {
	// Send 发送通知
	Send(ctx context.Context, message *Message) error
	// Test 测试通知
	Test(ctx context.Context) error
}

// Message 通知消息
type Message struct {
	Title   string                 // 标题
	Content string                 // 内容
	Level   NotificationLevel      // 级别：info, success, warning, error
	Extra   map[string]interface{} // 额外信息
}

// NotificationLevel 通知级别
type NotificationLevel string

const (
	LevelInfo    NotificationLevel = "info"
	LevelSuccess NotificationLevel = "success"
	LevelWarning NotificationLevel = "warning"
	LevelError   NotificationLevel = "error"
)

// BackupNotification 备份通知内容
type BackupNotification struct {
	TaskName     string
	HostName     string
	Databases    []string
	BackupType   string
	Status       string
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	FileSize     int64
	ErrorMessage string
}

// ToMessage 转换为通知消息
func (bn *BackupNotification) ToMessage() *Message {
	var level NotificationLevel
	var title string

	if bn.Status == "success" {
		level = LevelSuccess
		title = fmt.Sprintf("[成功] 备份任务 %s 完成", bn.TaskName)
	} else {
		level = LevelError
		title = fmt.Sprintf("[失败] 备份任务 %s 失败", bn.TaskName)
	}

	content := fmt.Sprintf(`
任务名称：%s
主机名称：%s
备份类型：%s
备份状态：%s
开始时间：%s
结束时间：%s
备份时长：%s
文件大小：%d MB
`,
		bn.TaskName,
		bn.HostName,
		bn.BackupType,
		bn.Status,
		bn.StartTime.Format("2006-01-02 15:04:05"),
		bn.EndTime.Format("2006-01-02 15:04:05"),
		bn.Duration.String(),
		bn.FileSize/1024/1024,
	)

	if bn.ErrorMessage != "" {
		content += fmt.Sprintf("\n错误信息：%s", bn.ErrorMessage)
	}

	return &Message{
		Title:   title,
		Content: content,
		Level:   level,
		Extra: map[string]interface{}{
			"task_name":  bn.TaskName,
			"host_name":  bn.HostName,
			"status":     bn.Status,
			"start_time": bn.StartTime,
		},
	}
}

// NewNotifier 创建通知实例
func NewNotifier(notifType string, config map[string]interface{}) (Notifier, error) {
	switch notifType {
	case "email":
		return NewEmailNotifier(config)
	case "webhook":
		// webhook类型需要根据webhook_type字段来判断具体类型
		webhookType, _ := config["webhook_type"].(string)
		switch webhookType {
		case "dingtalk":
			return NewDingTalkNotifier(config)
		case "wecom":
			return NewWeComNotifier(config)
		case "feishu", "slack", "custom":
			// 飞书、Slack和自定义webhook使用通用webhook处理
			return NewWebhookNotifier(config)
		default:
			return nil, fmt.Errorf("unsupported webhook type: %s", webhookType)
		}
	case "dingtalk":
		return NewDingTalkNotifier(config)
	case "wecom":
		return NewWeComNotifier(config)
	default:
		return nil, fmt.Errorf("unsupported notification type: %s", notifType)
	}
}
