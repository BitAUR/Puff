package auth

import (
	"Puff/internal/config"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		// 更新会话，刷新过期时间
		session.Set("user", user)
		session.Save()
		c.Next()
	}
}

func LoginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 使用 LoadConfig 获取最新的配置
	cfg, err := config.LoadConfig()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"title":   "登录",
			"content": "login",
			"error":   "无法加载配置",
		})
		return
	}

	if username == cfg.AuthUsername && password == cfg.AuthPassword {
		session := sessions.Default(c)
		session.Set("user", username)
		session.Save()
		c.Redirect(http.StatusFound, "/")
	} else {
		c.HTML(http.StatusUnauthorized, "layout.html", gin.H{
			"title":   "登录",
			"content": "login",
			"error":   "用户名或密码错误",
		})
	}
}

func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
