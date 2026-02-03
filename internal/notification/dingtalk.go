package notification

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DingTalkNotifier 钉钉通知
type DingTalkNotifier struct {
	config *DingTalkConfig
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"` // 签名密钥
}

// NewDingTalkNotifier 创建钉钉通知实例
func NewDingTalkNotifier(config map[string]interface{}) (*DingTalkNotifier, error) {
	webhookURL, _ := config["webhook_url"].(string)
	secret, _ := config["secret"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required")
	}

	return &DingTalkNotifier{
		config: &DingTalkConfig{
			WebhookURL: webhookURL,
			Secret:     secret,
		},
	}, nil
}

func (n *DingTalkNotifier) Send(ctx context.Context, message *Message) error {
	// 构建请求URL（如果有签名）
	requestURL := n.config.WebhookURL
	if n.config.Secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := n.generateSign(timestamp)
		requestURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", n.config.WebhookURL, timestamp, sign)
	}

	// 构建消息体
	content := fmt.Sprintf("## %s\n\n%s", message.Title, message.Content)
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": message.Title,
			"text":  content,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
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
		return fmt.Errorf("dingtalk error: %v", result["errmsg"])
	}

	return nil
}

func (n *DingTalkNotifier) Test(ctx context.Context) error {
	testMessage := &Message{
		Title:   "测试通知",
		Content: "这是一条测试通知，用于验证钉钉机器人配置是否正确。",
		Level:   LevelInfo,
	}

	return n.Send(ctx, testMessage)
}

// generateSign 生成签名
func (n *DingTalkNotifier) generateSign(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, n.config.Secret)
	h := hmac.New(sha256.New, []byte(n.config.Secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(signature)
}
