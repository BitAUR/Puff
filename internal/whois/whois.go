package whois

import (
	"io"
	"net"
	"strings"
	"time"
)

func QueryDomain(domain, whoisServer string) (bool, error) {
	conn, err := net.DialTimeout("tcp", whoisServer+":43", 10*time.Second)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	conn.Write([]byte(domain + "\r\n"))

	var response strings.Builder
	io.Copy(&response, conn)

	return !strings.Contains(strings.ToLower(response.String()), "no match"), nil
}

func GetTLD(domain string) string {
	parts := strings.Split(domain, ".")
	return parts[len(parts)-1]
}
