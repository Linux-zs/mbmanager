package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookNotifier 通用Webhook通知
type WebhookNotifier struct {
	config *WebhookConfig
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	WebhookURL  string `json:"webhook_url"`
	WebhookType string `json:"webhook_type"` // feishu, slack, custom
}

// NewWebhookNotifier 创建通用Webhook通知实例
func NewWebhookNotifier(config map[string]interface{}) (*WebhookNotifier, error) {
	webhookURL, _ := config["webhook_url"].(string)
	webhookType, _ := config["webhook_type"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required")
	}

	return &WebhookNotifier{
		config: &WebhookConfig{
			WebhookURL:  webhookURL,
			WebhookType: webhookType,
		},
	}, nil
}

func (n *WebhookNotifier) Send(ctx context.Context, message *Message) error {
	var payload map[string]interface{}

	// 根据不同的webhook类型构建不同的消息格式
	switch n.config.WebhookType {
	case "feishu":
		// 飞书消息格式
		content := fmt.Sprintf("**%s**\n\n%s", message.Title, message.Content)
		payload = map[string]interface{}{
			"msg_type": "text",
			"content": map[string]interface{}{
				"text": content,
			},
		}
	case "slack":
		// Slack消息格式
		payload = map[string]interface{}{
			"text": fmt.Sprintf("*%s*\n%s", message.Title, message.Content),
		}
	case "custom":
		// 自定义格式 - 使用通用JSON格式
		payload = map[string]interface{}{
			"title":   message.Title,
			"content": message.Content,
			"level":   string(message.Level),
			"extra":   message.Extra,
		}
	default:
		return fmt.Errorf("unsupported webhook type: %s", n.config.WebhookType)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", n.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (n *WebhookNotifier) Test(ctx context.Context) error {
	testMessage := &Message{
		Title:   "测试通知",
		Content: "这是一条测试通知，用于验证Webhook配置是否正确。",
		Level:   LevelInfo,
	}

	return n.Send(ctx, testMessage)
}
