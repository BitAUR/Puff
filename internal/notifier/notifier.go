package notifier

import (
	"Puff/internal/config"
	"crypto/tls"
	"fmt"
	"log"
	"net"
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

	// 创建邮件内容
	to := []string{cfg.RecipientEmail}
	subject := "域名状态变更提醒"
	body := generateEmailBody(notifications)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", cfg.SMTPUsername, cfg.RecipientEmail, subject, body))

	// 根据端口选择不同的发送方式
	var err error
	switch cfg.SMTPPort {
	case 25:
		err = sendMailInsecure(cfg, to, msg)
	case 465:
		err = sendMailSSL(cfg, to, msg)
	default: // 包括 587 端口
		err = sendMailTLS(cfg, to, msg)
	}

	if err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Println("邮件发送成功")
	return nil
}

func sendMailInsecure(cfg *config.Config, to []string, msg []byte) error {
	return smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPServer, cfg.SMTPPort),
		nil,
		cfg.SMTPUsername,
		to,
		msg,
	)
}

func sendMailSSL(cfg *config.Config, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", cfg.SMTPServer, cfg.SMTPPort), &tls.Config{
		ServerName: cfg.SMTPServer,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, cfg.SMTPServer)
	if err != nil {
		return err
	}
	defer c.Quit()

	return sendMail(c, cfg, to, msg)
}

func sendMailTLS(cfg *config.Config, to []string, msg []byte) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", cfg.SMTPServer, cfg.SMTPPort))
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, cfg.SMTPServer)
	if err != nil {
		return err
	}
	defer c.Quit()

	if err = c.StartTLS(&tls.Config{ServerName: cfg.SMTPServer}); err != nil {
		return err
	}

	return sendMail(c, cfg, to, msg)
}

func sendMail(c *smtp.Client, cfg *config.Config, to []string, msg []byte) error {
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPServer)
	if err := c.Auth(auth); err != nil {
		return err
	}

	if err := c.Mail(cfg.SMTPUsername); err != nil {
		return err
	}

	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

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
