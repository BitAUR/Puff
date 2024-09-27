package web

import (
	"Puff/internal/auth"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func StartServer() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 设置 session 中间件
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
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

	return r.Run(":" + os.Getenv("WEB_PORT"))
}

func handleLogin(c *gin.Context) {
	c.HTML(200, "layout.html", gin.H{
		"title":   "登录",
		"content": "login",
	})
}
