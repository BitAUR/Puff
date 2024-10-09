package web

import (
	"Puff/internal/config"
	"Puff/internal/monitor"
	"Puff/internal/notifier"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func handleIndex(c *gin.Context) {
	domains, err := config.LoadDomainList()
	if err != nil {
		log.Printf("加载域名错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	statuses := monitor.GetDomainStatuses()
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":    "域名管理",
		"Domains":  domains,
		"Statuses": statuses,
		"content":  "index",
	})
}
func handleGetDomains(c *gin.Context) {
	domains, err := config.LoadDomainList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, domains)
}

func handleAddDomain(c *gin.Context) {
	domain := c.PostForm("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "域名不能为空"})
		return
	}

	if err := config.AddDomain(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 重新加载域名列表
	domains, err := config.LoadDomainList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 更新监控系统中的域名列表
	monitor.UpdateDomainList(domains)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func handleDeleteDomain(c *gin.Context) {
	domain := c.Param("domain")
	if err := config.DeleteDomain(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 重新加载域名列表
	domains, err := config.LoadDomainList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 更新监控系统中的域名列表
	monitor.UpdateDomainList(domains)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func handleGetWhoisServers(c *gin.Context) {
	whoisServers, err := config.LoadWhoisServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, whoisServers)
}

func handleAddWhoisServer(c *gin.Context) {
	tld := c.PostForm("tld")
	server := c.PostForm("server")

	if tld == "" || server == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "TLD 和服务器地址都不能为空"})
		return
	}

	if err := config.AddWhoisServer(tld, server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func handleDeleteWhoisServer(c *gin.Context) {
	tld := c.Param("tld")

	if tld == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "TLD 不能为空"})
		return
	}

	if err := config.DeleteWhoisServer(tld); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func handleGetRecipientEmail(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"email": cfg.RecipientEmail})
}

func handleUpdateRecipientEmail(c *gin.Context) {
	var newEmail struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&newEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.UpdateRecipientEmail(newEmail.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recipient email updated successfully"})
}

func handleGetDomainStatuses(c *gin.Context) {
	statuses := monitor.GetDomainStatuses()
	c.JSON(http.StatusOK, statuses)
}

func handleRefreshStatuses(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	domains, err := config.LoadDomainList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	whoisServers, err := config.LoadWhoisServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monitor.RefreshAllDomains(domains, whoisServers, cfg)
	statuses := monitor.GetDomainStatuses()
	c.JSON(http.StatusOK, statuses)
}

func handleWhoisServers(c *gin.Context) {
	whoisServers, err := config.LoadWhoisServers()
	if err != nil {
		log.Printf("加载 Whois 服务器错误: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":        "Whois 服务器管理",
		"WhoisServers": whoisServers,
		"content":      "whois_servers",
	})

}

func handleSettings(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("加载配置出错: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":   "设置",
		"content": "settings",
		"config":  cfg,
	})
}

func handleUpdateSettings(c *gin.Context) {
	// 打印原始请求体
	body, _ := ioutil.ReadAll(c.Request.Body)
	log.Printf("原始请求体: %s", string(body))
	// 重新设置请求体，因为它已经被读取
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	var newConfig config.Config
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		log.Printf("绑定请求时出错: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("收到来自客户端的新配置: %+v", newConfig)

	// 直接使用新接收到的配置
	if err := config.SaveConfig(&newConfig); err != nil {
		log.Printf("保存配置时出错: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("配置保存成功")

	if err := config.ReloadConfig(); err != nil {
		log.Printf("重新加载配置时出错: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("配置重新加载成功")

	// 重启监控
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("监控重启时出现恐慌: %v", r)
			}
		}()
		whoisServers, err := config.LoadWhoisServers()
		if err != nil {
			log.Printf("加载 Whois 服务器时出错: %v", err)
			return
		}
		monitor.StartMonitoring(whoisServers, &newConfig)
	}()

	c.JSON(http.StatusOK, gin.H{"message": "设置已更新并重新加载"})
}

func handleAPISettings(c *gin.Context) {
	if c.Request.Method == "GET" {
		cfg, err := config.LoadConfig() // 每次请求都加载最新配置
		if err != nil {
			log.Printf("加载配置时出错: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "当前配置",
			"config": gin.H{
				"RECIPIENT_EMAIL":         cfg.RecipientEmail,
				"SMTP_SERVER":             cfg.SMTPServer,
				"SMTP_PORT":               cfg.SMTPPort,
				"SMTP_USERNAME":           cfg.SMTPUsername,
				"SMTP_PASSWORD":           cfg.SMTPPassword,
				"WEB_PORT":                cfg.WebPort,
				"AUTH_USERNAME":           cfg.AuthUsername,
				"AUTH_PASSWORD":           cfg.AuthPassword,
				"QUERY_FREQUENCY_SECONDS": cfg.QueryFrequencySeconds,
				"SESSION_SECRET":          cfg.SessionSecret,
			},
		})
	} else if c.Request.Method == "POST" {
		var newConfig config.Config
		if err := c.ShouldBindJSON(&newConfig); err != nil {
			c.HTML(http.StatusUnauthorized, "layout.html", gin.H{
				"error": "绑定请求时出错:" + err.Error(),
			})
			return
		}

		if err := config.SaveConfig(&newConfig); err != nil {
			c.HTML(http.StatusUnauthorized, "layout.html", gin.H{
				"error": "保存配置时出错:" + err.Error(),
			})
			return
		}

		log.Printf("新配置保存成功: %+v", newConfig)

		// 重启监控（如果需要）
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("监控重启时出现错误: %v", r)
				}
			}()
			whoisServers, err := config.LoadWhoisServers()
			if err != nil {
				log.Printf("加载 Whois 服务器时出错: %v", err)
				return
			}
			monitor.StartMonitoring(whoisServers, &newConfig)
		}()

		c.JSON(http.StatusOK, gin.H{
			"message": "设置已更新并重新加载",
			"config":  newConfig,
		})
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "不允许的方法"})
	}
}

func handleTestEmail(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "加载配置失败: " + err.Error()})
		return
	}

	testNotification := []notifier.DomainNotification{
		{
			Domain:        "example.com",
			IsFinalNotice: false,
			Status:        "测试状态",
		},
	}

	err = notifier.SendNotification(testNotification, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "发送测试邮件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "测试邮件发送成功"})
}

type GithubRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
}

func handleCheckUpdate(c *gin.Context) {
	currentVersion := "v0.2.2" // 当前版本

	// 获取 GitHub 最新 release
	resp, err := http.Get("https://api.bitaur.com/puff/version")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法检查更新"})
		return
	}
	defer resp.Body.Close()

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析更新信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currentVersion":  currentVersion,
		"latestVersion":   release.TagName,
		"publishedAt":     release.PublishedAt,
		"updateAvailable": release.TagName != currentVersion,
	})
}
