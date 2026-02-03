package notification

import (
	"context"
	"crypto/tls"
	"fmt"

	"gopkg.in/gomail.v2"
)

// EmailNotifier 邮件通知
type EmailNotifier struct {
	config *EmailConfig
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost string   `json:"smtp_host"`
	SMTPPort int      `json:"smtp_port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	UseSSL   bool     `json:"use_ssl"`
}

// NewEmailNotifier 创建邮件通知实例
func NewEmailNotifier(config map[string]interface{}) (*EmailNotifier, error) {
	smtpHost, _ := config["smtp_host"].(string)
	smtpPort, _ := config["smtp_port"].(float64)
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)
	from, _ := config["from"].(string)
	toList, _ := config["to"].([]interface{})
	useSSL, _ := config["use_ssl"].(bool)

	if smtpHost == "" {
		return nil, fmt.Errorf("smtp_host is required")
	}
	if smtpPort == 0 {
		smtpPort = 25
	}
	if from == "" {
		return nil, fmt.Errorf("from is required")
	}
	if len(toList) == 0 {
		return nil, fmt.Errorf("to is required")
	}

	// 转换收件人列表
	to := make([]string, 0, len(toList))
	for _, t := range toList {
		if email, ok := t.(string); ok {
			to = append(to, email)
		}
	}

	return &EmailNotifier{
		config: &EmailConfig{
			SMTPHost: smtpHost,
			SMTPPort: int(smtpPort),
			Username: username,
			Password: password,
			From:     from,
			To:       to,
			UseSSL:   useSSL,
		},
	}, nil
}

func (n *EmailNotifier) Send(ctx context.Context, message *Message) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.config.From)
	m.SetHeader("To", n.config.To...)
	m.SetHeader("Subject", message.Title)
	m.SetBody("text/plain", message.Content)

	d := gomail.NewDialer(n.config.SMTPHost, n.config.SMTPPort, n.config.Username, n.config.Password)

	// 根据端口和UseSSL配置选择加密方式
	if n.config.UseSSL {
		if n.config.SMTPPort == 465 {
			// 端口465使用SSL/TLS（直接加密连接）
			d.SSL = true
		} else {
			// 端口587等其他端口使用STARTTLS（先明文连接再升级）
			d.SSL = false
		}
		// 设置TLS配置（跳过证书验证）
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		// 不使用加密
		d.SSL = false
		d.TLSConfig = nil
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *EmailNotifier) Test(ctx context.Context) error {
	testMessage := &Message{
		Title:   "测试邮件",
		Content: "这是一封测试邮件，用于验证邮件配置是否正确。",
		Level:   LevelInfo,
	}

	return n.Send(ctx, testMessage)
}
