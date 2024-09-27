package main

import (
	"Puff/internal/config"
	"Puff/internal/monitor"
	"Puff/internal/web"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	// 确保配置文件存在
	if err := config.EnsureConfigFiles(); err != nil {
		log.Fatalf("确保配置文件存在时出错: %v", err)
	}

	// 获取 .env 文件的路径
	envPath := config.GetConfigPath(".env")

	// 加载 .env 文件
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("警告: 无法加载 .env 文件 (%s): %v", envPath, err)
	}
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 检查必要的环境变量
	requiredEnvVars := []string{"AUTH_USERNAME", "AUTH_PASSWORD", "SESSION_SECRET"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s 必须在 .env 文件中设置", envVar)
		}
	}

	// 加载 Whois 服务器信息
	whoisServers, err := config.LoadWhoisServers()
	if err != nil {
		log.Fatalf("加载 Whois 服务器失败: %v", err)
	}

	// 启动域名监控
	go func() {
		monitor.StartMonitoring(whoisServers, cfg)
	}()

	// 启动 Web 服务器
	go func() {
		if err := web.StartServer(); err != nil {
			log.Fatalf("启动 Web 服务器失败: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
