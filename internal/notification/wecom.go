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

// WeComNotifier 企业微信通知
type WeComNotifier struct {
	config *WeComConfig
}

// WeComConfig 企业微信配置
type WeComConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// NewWeComNotifier 创建企业微信通知实例
func NewWeComNotifier(config map[string]interface{}) (*WeComNotifier, error) {
	webhookURL, _ := config["webhook_url"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required")
	}

	return &WeComNotifier{
		config: &WeComConfig{
			WebhookURL: webhookURL,
		},
	}, nil
}

func (n *WeComNotifier) Send(ctx context.Context, message *Message) error {
	// 构建消息体
	content := fmt.Sprintf("**%s**\n\n%s", message.Title, message.Content)
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": content,
		},
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

	// 检查响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("wecom error: %v", result["errmsg"])
	}

	return nil
}

func (n *WeComNotifier) Test(ctx context.Context) error {
	testMessage := &Message{
		Title:   "测试通知",
		Content: "这是一条测试通知，用于验证企业微信机器人配置是否正确。",
		Level:   LevelInfo,
	}

	return n.Send(ctx, testMessage)
}
