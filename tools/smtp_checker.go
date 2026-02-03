package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

func main() {
	// 163邮箱配置
	smtpHost := "smtp.163.com"
	smtpPort := 587 // 可以尝试 465
	username := "dtm0527@163.com"
	password := "HKcGE34UKSkh3GEq" // 授权码
	from := "dtm0527@163.com"
	to := "hank3997@163.com" // 修改为实际的收件人地址

	// 如果通过命令行参数指定收件人
	if len(os.Args) > 1 {
		to = os.Args[1]
	}

	fmt.Printf("测试SMTP连接...\n")
	fmt.Printf("服务器: %s:%d\n", smtpHost, smtpPort)
	fmt.Printf("用户名: %s\n", username)
	fmt.Printf("发件人: %s\n", from)
	fmt.Printf("收件人: %s\n", to)
	fmt.Println("---")

	// 创建邮件
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "SMTP测试邮件")
	m.SetBody("text/plain", "这是一封测试邮件，用于验证SMTP配置是否正确。\n\n如果您收到这封邮件，说明配置成功！")

	// 测试方案1：端口587 + STARTTLS
	fmt.Println("方案1: 端口587 + STARTTLS")
	d1 := gomail.NewDialer(smtpHost, 587, username, password)
	d1.SSL = false
	d1.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d1.DialAndSend(m); err != nil {
		fmt.Printf("❌ 失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 成功！邮件已发送\n")
		return
	}

	// 测试方案2：端口465 + SSL
	fmt.Println("方案2: 端口465 + SSL")
	d2 := gomail.NewDialer(smtpHost, 465, username, password)
	d2.SSL = true
	d2.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d2.DialAndSend(m); err != nil {
		fmt.Printf("❌ 失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 成功！邮件已发送\n")
		return
	}

	// 测试方案3：端口587 + 不设置TLSConfig
	fmt.Println("方案3: 端口587 + 默认TLS配置")
	d3 := gomail.NewDialer(smtpHost, 587, username, password)
	d3.SSL = false

	if err := d3.DialAndSend(m); err != nil {
		fmt.Printf("❌ 失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 成功！邮件已发送\n")
		return
	}

	// 测试方案4：端口465 + 不设置TLSConfig
	fmt.Println("方案4: 端口465 + 默认TLS配置")
	d4 := gomail.NewDialer(smtpHost, 465, username, password)
	d4.SSL = true

	if err := d4.DialAndSend(m); err != nil {
		fmt.Printf("❌ 失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 成功！邮件已发送\n")
		return
	}

	fmt.Println("所有方案都失败了，请检查：")
	fmt.Println("1. 网络连接是否正常")
	fmt.Println("2. 授权码是否正确")
	fmt.Println("3. 用户名和发件人是否一致")
	fmt.Println("4. 防火墙是否阻止了587或465端口")

	log.Fatal("SMTP测试失败")
}
