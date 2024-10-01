package web

import (
	"Puff/internal/auth"
	"Puff/internal/config"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func StartServer() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 使用 LoadConfig 获取最新的配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// 设置 session 中间件
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		MaxAge:   int(2 * time.Hour.Seconds()), // 设置会话最大存活时间为2小时
		Path:     "/",
		Secure:   false, // 如果使用HTTPS，设置为true
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	// 公开路由
	r.GET("/login", handleLogin)
	r.POST("/login", auth.LoginHandler)
	r.GET("/logout", auth.LogoutHandler)

	// 受保护的路由
	authorized := r.Group("/")
	authorized.Use(auth.AuthMiddleware())
	{
		authorized.GET("/", handleIndex)
		authorized.GET("/domains", handleIndex)
		authorized.GET("/whois-servers", handleWhoisServers)

		// API 路由
		authorized.POST("/domains", handleAddDomain)
		authorized.DELETE("/domains/:domain", handleDeleteDomain)
		authorized.POST("/whois-servers", handleAddWhoisServer)
		authorized.DELETE("/whois-servers/:tld", handleDeleteWhoisServer)
		authorized.GET("/recipient-email", handleGetRecipientEmail)
		authorized.POST("/recipient-email", handleUpdateRecipientEmail)
		authorized.GET("/domain-statuses", handleGetDomainStatuses)
		authorized.POST("/refresh-statuses", handleRefreshStatuses)

		authorized.GET("/api/domains", handleGetDomains)
		authorized.GET("/api/whois-servers", handleGetWhoisServers)

		authorized.GET("/settings", handleSettings)
		authorized.POST("/settings", handleUpdateSettings)
		authorized.GET("/api/settings", handleAPISettings)
		authorized.POST("/api/settings", handleAPISettings)
	}

	return r.Run(":" + strconv.Itoa(cfg.WebPort))
}

func handleLogin(c *gin.Context) {
	c.HTML(200, "layout.html", gin.H{
		"title":   "登录",
		"content": "login",
	})
}
