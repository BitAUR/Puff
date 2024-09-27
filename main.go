package main

import (
	"Puff/internal/config"
	"Puff/internal/monitor"
	"Puff/internal/web"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 检查必要的环境变量
	requiredEnvVars := []string{"AUTH_USERNAME", "AUTH_PASSWORD", "SESSION_SECRET"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s must be set in .env file", envVar)
		}
	}

	// 加载配置S
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 加载域名列表和 Whois 服务器信息
	whoisServers, err := config.LoadWhoisServers()
	if err != nil {
		log.Fatalf("Failed to load Whois servers: %v", err)
	}

	// 启动 Web 服务器
	go func() {
		if err := web.StartServer(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// 启动域名监控
	monitor.StartMonitoring(whoisServers, cfg)
}
