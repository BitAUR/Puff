package monitor

import (
	"Puff/internal/config"
	"Puff/internal/notifier"
	"Puff/internal/whois"
	"log"
	"sync"
	"time"
)

type DomainStatus struct {
	Domain            string
	Registered        bool
	Redemption        bool
	PendingDelete     bool
	ExpirationDate    time.Time
	LastChecked       time.Time
	FirstNotifiedAt   time.Time
	CheckCount        int
	NeedsNotification bool
	IsFinalNotice     bool
	FinalNoticed      bool // 添加这个字段
}

var (
	domainStatuses   = make(map[string]*DomainStatus)
	statusMutex      sync.RWMutex
	availableDomains []string
)

var (
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex // 添加互斥锁
)

func updateDomainStatus(domain string, registered bool) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	status, exists := domainStatuses[domain]
	if !exists {
		status = &DomainStatus{Domain: domain}
		domainStatuses[domain] = status
	}

	// 如果已经发送了最终通知，不做任何操作
	if status.FinalNoticed {
		return
	}

	status.Registered = registered
	status.LastChecked = time.Now()

	if !registered {
		if status.FirstNotifiedAt.IsZero() {
			log.Printf("域名 %s 首次被检测为可注册", domain)
			status.FirstNotifiedAt = time.Now()
			status.CheckCount = 1
			status.NeedsNotification = true
			status.IsFinalNotice = false
		} else {
			status.CheckCount++
			log.Printf("域名 %s 仍然可注册。检查次数：%d", domain, status.CheckCount)
			if status.CheckCount == 3 {
				log.Printf("域名 %s 已连续三次检测为可注册。准备发送最终通知，且后续不再检查。", domain)
				status.NeedsNotification = true
				status.IsFinalNotice = true
				status.FinalNoticed = true // 设置为最终通知已发送
			} else {
				status.NeedsNotification = false
			}
		}
	} else {
		// 如果域名变为已注册，重置状态
		log.Printf("域名 %s 已被注册。重置状态。", domain)
		status.FirstNotifiedAt = time.Time{}
		status.CheckCount = 0
		status.FinalNoticed = false
		status.NeedsNotification = false
		status.IsFinalNotice = false
	}

	if status.NeedsNotification {
		log.Printf("将域名 %s 添加到可用域名列表", domain)
		availableDomains = append(availableDomains, domain)
	}
}

func StartMonitoring(whoisServers map[string]string, cfg *config.Config) {
	startTime := time.Now()
	log.Printf("开始域名检查，时间：%s", startTime.Format("2006-01-02 15:04:05"))

	domains, err := config.LoadDomainList()
	if err != nil {
		log.Printf("加载域名列表失败: %v", err)
		return
	}

	RefreshAllDomains(domains, whoisServers, cfg)

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("域名检查完成，时间：%s，耗时：%v", endTime.Format("2006-01-02 15:04:05"), duration)
}

func WaitForMonitoring() {
	wg.Wait()
}

func checkAllDomains(cfg *config.Config) {
	startTime := time.Now()
	log.Printf("开始域名检查，时间：%s", startTime.Format("2006-01-02 15:04:05"))

	// 重新加载域名列表
	domains, err := config.LoadDomainList()
	if err != nil {
		log.Printf("加载域名列表时出错：%v", err)
		return
	}

	// 更新监控系统中的域名列表
	UpdateDomainList(domains)

	whoisServers, err := config.LoadWhoisServers()
	if err != nil {
		log.Printf("加载 Whois 服务器列表时出错：%v", err)
		return
	}

	// 使用 whoisServers 检查所有域名
	for _, domain := range domains {
		checkDomain(domain, whoisServers, cfg)
	}

	var notifications []notifier.DomainNotification
	statusMutex.RLock()
	for _, status := range domainStatuses {
		if status.NeedsNotification && !status.Registered {
			notifications = append(notifications, notifier.DomainNotification{
				Domain:        status.Domain,
				IsFinalNotice: status.IsFinalNotice,
			})
		}
	}
	statusMutex.RUnlock()

	if len(notifications) > 0 {
		log.Printf("发现 %d 个需要通知的域名。准备发送通知。", len(notifications))
		if err := notifier.SendNotification(notifications, cfg); err != nil {
			log.Printf("发送通知时出错：%v", err)
		} else {
			log.Printf("已发送通知")
			resetNotificationFlags(notifications)
		}
	} else {
		log.Println("本次检查未发现需要通知的域名。")
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("域名检查完成，时间：%s，耗时：%v", endTime.Format("2006-01-02 15:04:05"), duration)
}

func RefreshAllDomains(domains []string, whoisServers map[string]string, cfg *config.Config) {
	var wg sync.WaitGroup
	results := make(chan whois.DomainStatus, len(domains))

	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			statusMutex.RLock()
			status, exists := domainStatuses[d]
			if exists && status.FinalNoticed { // 使用 FinalNoticed 而不是 FinalNotice
				statusMutex.RUnlock()
				return
			}
			statusMutex.RUnlock()
			result, err := checkDomain(d, whoisServers, cfg)
			if err != nil {
				log.Printf("检查域名 %s 错误: %v", d, err)
				return
			}
			results <- result
		}(domain)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	processResults(results, cfg)
}

type DomainCheckResult struct {
	Domain     string
	Registered bool
	Error      error
}

func processResults(results <-chan whois.DomainStatus, cfg *config.Config) {
	var notifications []notifier.DomainNotification

	for result := range results {
		statusMutex.Lock()
		status, exists := domainStatuses[result.Domain]
		if !exists {
			status = &DomainStatus{Domain: result.Domain}
			domainStatuses[result.Domain] = status
		}

		prevStatus := *status // 保存之前的状态

		status.Registered = result.Registered
		status.Redemption = result.Redemption
		status.PendingDelete = result.PendingDelete
		status.ExpirationDate = result.ExpirationDate
		status.LastChecked = time.Now()

		statusChanged := !status.Registered || status.Redemption || status.PendingDelete

		if statusChanged {
			if status.FirstNotifiedAt.IsZero() {
				status.FirstNotifiedAt = time.Now()
				status.CheckCount = 1
				status.NeedsNotification = true
				log.Printf("域名 %s 状态首次变化，将发送第一次通知", status.Domain)
			} else {
				status.CheckCount++
				log.Printf("域名 %s 状态变化次数: %d", status.Domain, status.CheckCount)
				if status.CheckCount == 3 {
					status.NeedsNotification = true
					status.IsFinalNotice = true
					log.Printf("域名 %s 将发送最终通知", status.Domain)
				} else {
					status.NeedsNotification = false
				}
			}
		} else if prevStatus.Registered != status.Registered ||
			prevStatus.Redemption != status.Redemption ||
			prevStatus.PendingDelete != status.PendingDelete {
			status.NeedsNotification = false
			status.IsFinalNotice = false
			status.FirstNotifiedAt = time.Time{}
			status.CheckCount = 0
		}

		if status.NeedsNotification {
			notifications = append(notifications, notifier.DomainNotification{
				Domain:        status.Domain,
				IsFinalNotice: status.IsFinalNotice,
				Status:        getDomainStatusString(status),
			})
		}

		statusMutex.Unlock()
	}

	if len(notifications) > 0 {
		if err := notifier.SendNotification(notifications, cfg); err != nil {
			log.Printf("发送邮件错误: %v", err)
		} else {
			resetNotificationFlags(notifications)
		}
	}
}

func checkDomain(domain string, whoisServers map[string]string, cfg *config.Config) (whois.DomainStatus, error) {
	tld := whois.GetTLD(domain)
	whoisServer, ok := whoisServers[tld]
	if !ok {
		log.Printf("未找到域名 %s 的 Whois 服务器", domain)
		return whois.DomainStatus{}, nil
	}

	status, err := whois.QueryDomain(domain, whoisServer)
	if err != nil {
		log.Printf("查询域名 %s 时出错：%v", domain, err)
		return whois.DomainStatus{}, err
	}

	logDomainStatus(domain, status)
	return status, nil
}

func logDomainStatus(domain string, status whois.DomainStatus) {
	if !status.Registered {
		log.Printf("域名 %s 状态: 可注册", domain)
	} else if status.PendingDelete {
		log.Printf("域名 %s 状态: 待删除", domain)
	} else if status.Redemption {
		log.Printf("域名 %s 状态: 赎回期", domain)
	} else {
		log.Printf("域名 %s 状态: 已注册", domain)
	}
}

func resetNotificationFlags(notifications []notifier.DomainNotification) {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	for _, n := range notifications {
		if status, exists := domainStatuses[n.Domain]; exists {
			status.NeedsNotification = false
			if status.IsFinalNotice {
				status.FirstNotifiedAt = time.Time{}
				status.CheckCount = 0
				status.IsFinalNotice = false
			}
		}
	}
}

func GetDomainStatuses() []DomainStatus {
	statusMutex.RLock()
	defer statusMutex.RUnlock()
	statuses := make([]DomainStatus, 0, len(domainStatuses))
	for _, status := range domainStatuses {
		statuses = append(statuses, *status)
	}
	return statuses
}

func UpdateDomainList(domains []string) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	// 删除不再监控的域名
	for domain := range domainStatuses {
		if !contains(domains, domain) {
			delete(domainStatuses, domain)
		}
	}

	// 添加新的域名
	for _, domain := range domains {
		if _, exists := domainStatuses[domain]; !exists {
			domainStatuses[domain] = &DomainStatus{Domain: domain}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func getDomainStatusString(status *DomainStatus) string {
	if !status.Registered {
		return "可注册"
	} else if status.PendingDelete {
		return "待删除"
	} else if status.Redemption {
		return "赎回期"
	} else {
		return "已注册"
	}
}
