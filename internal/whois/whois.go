package whois

import (
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
		Domain: domain,
	}

	// 检查域名是否注册
	status.Registered = !containsAny(responseLower, notRegisteredPhrases())

	// 检查域名是否处于赎回期
	status.Redemption = containsAny(responseLower, redemptionPhrases())

	// 检查域名是否处于待删除状态
	status.PendingDelete = containsAny(responseLower, pendingDeletePhrases())

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
