package whois

import (
	"io"
	"net"
	"strings"
	"time"
)

type DomainStatus struct {
	Domain         string
	Registered     bool
	Expired        bool
	Redemption     bool
	PendingDelete  bool
	ExpirationDate time.Time
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

// 保留原有的函数
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

// 新增函数
func redemptionPhrases() []string {
	return []string{
		"redemption period",
		"redemptionperiod",
		"status: redemption",
	}
}

// 新增函数
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
	parts := strings.Split(domain, ".")
	return parts[len(parts)-1]
}
