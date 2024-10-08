package notifier

import (
	"Puff/internal/config"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

type DomainNotification struct {
	Domain        string
	IsFinalNotice bool
	Status        string
}

func SendNotification(notifications []DomainNotification, cfg *config.Config) error {
	log.Printf("开始发送邮件通知")

	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPServer)

	to := []string{cfg.RecipientEmail}
	subject := "域名状态变更提醒"

	if len(notifications) == 1 && notifications[0].Domain == "example.com" {
		subject = "测试邮件 - " + subject
	}

	body := generateEmailBody(notifications)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", cfg.SMTPUsername, cfg.RecipientEmail, subject, body))

	// 创建TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         cfg.SMTPServer,
	}

	// 连接到SMTP服务器
	conn, err := smtp.Dial(fmt.Sprintf("%s:%d", cfg.SMTPServer, cfg.SMTPPort))
	if err != nil {
		return fmt.Errorf("连接到SMTP服务器失败: %v", err)
	}
	defer conn.Close()

	// 尝试启用TLS
	if err = conn.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("启用TLS失败: %v", err)
	}

	// 进行身份验证
	if err = conn.Auth(auth); err != nil {
		return fmt.Errorf("SMTP身份验证失败: %v", err)
	}

	// 设置发件人
	if err = conn.Mail(cfg.SMTPUsername); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	for _, addr := range to {
		if err = conn.Rcpt(addr); err != nil {
			return fmt.Errorf("设置收件人失败: %v", err)
		}
	}

	// 发送邮件内容
	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("准备发送邮件内容失败: %v", err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("完成邮件内容写入失败: %v", err)
	}

	log.Println("邮件发送成功")
	return nil
}

func generateEmailBody(notifications []DomainNotification) string {
	var body strings.Builder

	body.WriteString(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { width: 100%; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #161616; color: white; padding: 10px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { text-align: center; font-size: 0.8em; color: #777; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>域名状态变更提醒</h1>
        </div>
        <div class="content">
            <p>尊敬的用户，</p>
            <p>以下域名的状态发生了变化：</p>
            <ul>
    `)

	for _, n := range notifications {
		body.WriteString(fmt.Sprintf("<li>%s: %s", n.Domain, n.Status))
		if n.IsFinalNotice {
			body.WriteString(" (最终通知)")
		}
		body.WriteString("</li>")
	}

	body.WriteString(fmt.Sprintf(`
            </ul>
            <p>如果您对这些域名感兴趣，请尽快采取相应的行动。</p>
            <p>检测时间：%s</p>
        </div>
        <div class="footer">
            <p>此邮件由 Puff 自动发送，请勿直接回复。</p>
        </div>
    </div>
</body>
</html>
    `, time.Now().Format("2006年01月02日 15:04:05")))

	return body.String()
}
