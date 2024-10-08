package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

var configDir string

func init() {
	// 检查环境变量中是否指定了配置目录
	if envConfigDir := os.Getenv("CONFIG_DIR"); envConfigDir != "" {
		configDir = envConfigDir
	} else {
		// 如果环境变量未设置，则使用默认值
		configDir = "./data"
	}

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("无法创建配置目录: %v", err)
	}
}

// GetConfigDir 返回配置目录的路径
func GetConfigDir() string {
	return configDir
}

type Config struct {
	SMTPServer            string `json:"SMTP_SERVER"`
	SMTPPort              int    `json:"SMTP_PORT"`
	SMTPUsername          string `json:"SMTP_USERNAME"`
	SMTPPassword          string `json:"SMTP_PASSWORD"`
	RecipientEmail        string `json:"RECIPIENT_EMAIL"`
	WebPort               int    `json:"WEB_PORT"`
	AuthUsername          string `json:"AUTH_USERNAME"`
	AuthPassword          string `json:"AUTH_PASSWORD"`
	SessionSecret         string `json:"SESSION_SECRET"`
	QueryFrequencySeconds int    `json:"QUERY_FREQUENCY_SECONDS"`
}

func ensureConfigFiles() error {
	files := map[string]string{
		".env": `RECIPIENT_EMAIL="mail@yourdomain.com"
SMTP_PASSWORD="you_password"
SMTP_PORT=587
SMTP_SERVER="smtp.qq.com"
SMTP_USERNAME="mail@yourdomain.com"
WEB_PORT=8080
AUTH_USERNAME="admin"
AUTH_PASSWORD="admin"
SESSION_SECRET="your_random_secret_string"
QUERY_FREQUENCY_SECONDS=300
`,
		"list.yml": `domains: []
`,
		"whois.yml": `whois_servers:
  cn: whois.cnnic.cn
  com: whois.verisign-grs.com
  net: whois.verisign-grs.com
  org: whois.pir.org
`,
	}

	for file, content := range files {
		filePath := filepath.Join(configDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("配置文件 %s 不存在，正在创建...", file)
			err := os.WriteFile(filePath, []byte(content), 0644)
			if err != nil {
				return err
			}
			log.Printf("已创建配置文件 %s 并写入默认内容", file)
		}
	}
	return nil
}

func getConfigPath(filename string) string {
	dir := "data"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("创建配置目录: %s", dir)
		os.Mkdir(dir, 0755)
	}
	path := filepath.Join(dir, filename)
	return path
}

func LoadConfig() (*Config, error) {
	if err := ensureConfigFiles(); err != nil {
		return nil, err
	}

	envPath := getConfigPath(".env")

	// 读取 .env 文件
	envMap, err := godotenv.Read(envPath)
	if err != nil {
		log.Printf("警告: 无法读取 .env 文件 (%s): %v", envPath, err)
		// 继续执行，使用当前的环境变量
		envMap = make(map[string]string)
	}

	// 使用读取的值，如果没有则回退到环境变量
	getEnv := func(key string) string {
		if value, exists := envMap[key]; exists {
			return value
		}
		return os.Getenv(key)
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT"))
	webPort, _ := strconv.Atoi(getEnv("WEB_PORT"))

	QueryFrequencySeconds, _ := strconv.Atoi(getEnv("QUERY_FREQUENCY_SECONDS"))
	if QueryFrequencySeconds == 0 {
		QueryFrequencySeconds = 300 // 默认为5分钟
	}

	config := &Config{
		SMTPServer:            getEnv("SMTP_SERVER"),
		SMTPPort:              smtpPort,
		SMTPUsername:          getEnv("SMTP_USERNAME"),
		SMTPPassword:          getEnv("SMTP_PASSWORD"),
		RecipientEmail:        getEnv("RECIPIENT_EMAIL"),
		WebPort:               webPort,
		AuthUsername:          getEnv("AUTH_USERNAME"),
		AuthPassword:          getEnv("AUTH_PASSWORD"),
		SessionSecret:         getEnv("SESSION_SECRET"),
		QueryFrequencySeconds: QueryFrequencySeconds,
	}

	// 清理 envMap 以释放内存
	for k := range envMap {
		delete(envMap, k)
	}

	return config, nil
}

func LoadDomainList() ([]string, error) {
	file, err := os.ReadFile(getConfigPath("list.yml"))
	if err != nil {
		return nil, err
	}

	var data struct {
		Domains []string `yaml:"domains"`
	}

	if err := yaml.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	return data.Domains, nil
}

func LoadWhoisServers() (map[string]string, error) {
	file, err := os.ReadFile(getConfigPath("whois.yml"))
	if err != nil {
		return nil, err
	}

	var data struct {
		WhoisServers map[string]string `yaml:"whois_servers"`
	}

	if err := yaml.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	return data.WhoisServers, nil
}

func AddDomain(domain string) error {
	domains, err := LoadDomainList()
	if err != nil {
		return err
	}

	domains = append(domains, domain)
	return saveDomainList(domains)
}

func DeleteDomain(domain string) error {
	domains, err := LoadDomainList()
	if err != nil {
		return err
	}

	for i, d := range domains {
		if d == domain {
			domains = append(domains[:i], domains[i+1:]...)
			break
		}
	}

	return saveDomainList(domains)
}

func saveDomainList(domains []string) error {
	data := struct {
		Domains []string `yaml:"domains"`
	}{
		Domains: domains,
	}

	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath("list.yml"), yamlData, 0644)
}

func AddWhoisServer(tld, server string) error {
	whoisServers, err := LoadWhoisServers()
	if err != nil {
		return err
	}

	whoisServers[tld] = server
	return saveWhoisServers(whoisServers)
}

func DeleteWhoisServer(tld string) error {
	whoisServers, err := LoadWhoisServers()
	if err != nil {
		return err
	}

	delete(whoisServers, tld)
	return saveWhoisServers(whoisServers)
}

func saveWhoisServers(whoisServers map[string]string) error {
	data := struct {
		WhoisServers map[string]string `yaml:"whois_servers"`
	}{
		WhoisServers: whoisServers,
	}

	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath("whois.yml"), yamlData, 0644)
}

func UpdateRecipientEmail(email string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	cfg.RecipientEmail = email

	return saveConfig(cfg)
}

func saveConfig(cfg *Config) error {
	env, err := godotenv.Read(getConfigPath(".env"))
	if err != nil {
		return err
	}

	env["QUERY_FREQUENCY_SECONDS"] = strconv.Itoa(cfg.QueryFrequencySeconds)

	return godotenv.Write(env, getConfigPath(".env"))

}

// 将 ensureConfigFiles 改为公开函数
func EnsureConfigFiles() error {
	files := map[string]string{
		".env": `RECIPIENT_EMAIL="mail@yourdomain.com"
SMTP_PASSWORD="you_password"
SMTP_PORT=587
SMTP_SERVER="smtp.qq.com"
SMTP_USERNAME="mail@yourdomain.com"
WEB_PORT=8080
AUTH_USERNAME="admin"
AUTH_PASSWORD="admin"
SESSION_SECRET="your_random_secret_string"
`,
		"list.yml": `domains: []
`,
		"whois.yml": `whois_servers:
  cn: whois.cnnic.cn
  com: whois.verisign-grs.com
  net: whois.verisign-grs.com
  org: whois.pir.org
`,
	}

	for file, content := range files {
		filePath := filepath.Join(configDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("配置文件 %s 不存在，正在创建...", file)
			err := os.WriteFile(filePath, []byte(content), 0644)
			if err != nil {
				return err
			}
			log.Printf("已创建配置文件 %s 并写入默认内容", file)
		}
	}
	return nil
}

// 将 getConfigPath 改为公开函数
func GetConfigPath(filename string) string {
	return filepath.Join(configDir, filename)
}

func SaveConfig(cfg *Config) error {
	envPath := getConfigPath(".env")

	env, err := godotenv.Read(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println(".env 文件不存在，将创建新文件")
			env = make(map[string]string)
		} else {
			log.Printf("读取 .env 文件时出错: %v", err)
			return fmt.Errorf("error reading .env file: %w", err)
		}
	}

	// 辅助函数，只在新值非空时更新
	updateEnv := func(key string, value string) {
		if value != "" {
			env[key] = value
		}
	}

	updateEnv("SMTP_SERVER", cfg.SMTPServer)
	updateEnv("SMTP_USERNAME", cfg.SMTPUsername)
	updateEnv("SMTP_PASSWORD", cfg.SMTPPassword)
	updateEnv("RECIPIENT_EMAIL", cfg.RecipientEmail)
	updateEnv("AUTH_USERNAME", cfg.AuthUsername)
	updateEnv("AUTH_PASSWORD", cfg.AuthPassword)
	updateEnv("SESSION_SECRET", cfg.SessionSecret)

	// 对于数值类型，只在非零时更新
	if cfg.SMTPPort != 0 {
		env["SMTP_PORT"] = strconv.Itoa(cfg.SMTPPort)
	}
	if cfg.WebPort != 0 {
		env["WEB_PORT"] = strconv.Itoa(cfg.WebPort)
	}
	if cfg.QueryFrequencySeconds != 0 {
		env["QUERY_FREQUENCY_SECONDS"] = strconv.Itoa(cfg.QueryFrequencySeconds)
	}

	if err := godotenv.Write(env, envPath); err != nil {
		log.Printf("写入 .env 文件时出错: %v", err)
		return fmt.Errorf("error writing .env file: %w", err)
	}

	log.Println("配置已成功保存到 .env 文件")

	// 验证文件是否真的被写入
	content, err := ioutil.ReadFile(envPath)
	if err != nil {
		log.Printf("读取刚写入的 .env 文件时出错: %v", err)
	} else {
		log.Printf(".env 文件写入后的内容:\n%s", string(content))
	}

	return nil
}

var configMutex sync.RWMutex

func ReloadConfig() error {
	newConfig, err := LoadConfig()
	if err != nil {
		return err
	}

	configMutex.Lock()
	globalConfig = newConfig
	configMutex.Unlock()
	return nil
}

var globalConfig *Config

func init() {
	var err error
	globalConfig, err = LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}
