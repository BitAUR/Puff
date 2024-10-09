package whois

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

type DomainStatus struct {
	Domain         string
	Registered     bool
	Expired        bool
	Redemption     bool
	PendingDelete  bool
	ExpirationDate time.Time
	NoWhoisServer  bool
	RawWhois       string // 新增字段，存储原始 WHOIS 响应
}

func QueryDomain(domain, whoisServer string) (DomainStatus, error) {
	conn, err := net.DialTimeout("tcp", whoisServer+":43", 10*time.Second)
	if err != nil {
		return DomainStatus{}, err
	}
	defer conn.Close()

	conn.Write([]byte(domain + "\r\n"))

	var response strings.Builder
	io.Copy(&response, conn)

	responseStr := response.String()
	responseLower := strings.ToLower(responseStr)

	status := DomainStatus{
		Domain:   domain,
		RawWhois: responseStr, // 存储原始 WHOIS 响应
	}

	// 特殊处理 .cc 域名
	if strings.HasSuffix(domain, ".cc") {
		return handleCCDomain(responseLower, domain, responseStr)
	}

	// 检查域名是否处于待删除状态
	status.PendingDelete = containsAny(responseLower, pendingDeletePhrases())

	// 检查域名是否处于赎回期
	status.Redemption = containsAny(responseLower, redemptionPhrases())

	// 检查域名是否注册
	status.Registered = !containsAny(responseLower, notRegisteredPhrases()) || status.PendingDelete || status.Redemption

	// 输出查询到的 WHOIS 信息
	fmt.Printf("Domain: %s\nWHOIS Response:\n%s\n", domain, responseStr)

	return status, nil
}

func handleCCDomain(response, domain, rawWhois string) (DomainStatus, error) {
	status := DomainStatus{
		Domain:   domain,
		RawWhois: rawWhois, // 存储原始 WHOIS 响应
	}

	if strings.Contains(response, "no match") {
		// 域名未注册
		status.Registered = false
	} else if strings.Contains(response, "pendingdelete") {
		// 域名处于待删除状态，但仍然视为已注册
		status.Registered = true
		status.PendingDelete = true
	} else {
		// 其他情况视为已注册
		status.Registered = true
	}

	// 输出查询到的 WHOIS 信息
	fmt.Printf("Domain: %s\nWHOIS Response:\n%s\n", domain, rawWhois)

	return status, nil
}

// 未注册
func notRegisteredPhrases() []string {
	return []string{
		"no match",
		"no object found",
		"not found",
		"no entries found",
		"no data found",
		"domain not found",
		"not exist",
		"is available",
	}
}

// 赎回期
func redemptionPhrases() []string {
	return []string{
		"redemption period",
		"redemptionperiod",
		"status: redemption",
	}
}

// 正在删除
func pendingDeletePhrases() []string {
	return []string{
		"pending delete",
		"pendingdelete",
		"status: pending delete",
		"ICANN ",
	}
}

func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

func GetTLD(domain string) string {
	// 使用 publicsuffix 库获取有效的顶级域名
	suffix, _ := publicsuffix.PublicSuffix(domain)

	// 如果无法获取有效的顶级域名，则回退到原来的逻辑
	if suffix == "" {
		parts := strings.Split(domain, ".")
		return parts[len(parts)-1]
	}

	return suffix
}
