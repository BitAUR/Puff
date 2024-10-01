document.addEventListener('DOMContentLoaded', function() {
    // 域名管理
    const addDomainForm = document.getElementById('add-domain-form');
    const domainList = document.getElementById('domain-list');
    const refreshStatusBtn = document.getElementById('refresh-status-btn');

    if (addDomainForm) {
        addDomainForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const newDomain = document.getElementById('new-domain').value;
            addDomain(newDomain);
        });
    }

    if (domainList) {
        domainList.addEventListener('click', function(e) {
            if (e.target.classList.contains('delete-domain')) {
                const domain = e.target.getAttribute('data-domain');
                deleteDomain(domain);
            }
        });
    }

    if (refreshStatusBtn) {
        refreshStatusBtn.addEventListener('click', refreshDomainStatuses);
    }

    // Whois 服务器管理
    const addWhoisServerForm = document.getElementById('add-whois-server-form');
    const whoisServerList = document.getElementById('whois-server-list');

    if (addWhoisServerForm) {
        addWhoisServerForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const newTLD = document.getElementById('new-tld').value;
            const newServer = document.getElementById('new-server').value;
            addWhoisServer(newTLD, newServer);
        });
    }

    if (whoisServerList) {
        whoisServerList.addEventListener('click', function(e) {
            if (e.target.classList.contains('delete-whois-server')) {
                const tld = e.target.getAttribute('data-tld');
                deleteWhoisServer(tld);
            }
        });
    }

    if (window.location.pathname === '/' || window.location.pathname === '/domains') {
        loadDomains();
    } else if (window.location.pathname === '/whois-servers') {
        loadWhoisServers();
    }

    // 初始加载域名状态
    loadDomainStatuses();
    // 开始定时刷新
    setInterval(loadDomainStatuses, 60000); // 每分钟刷新一次

    const statusTable = document.getElementById('domain-status-list');
    if (statusTable) {
        statusTable.addEventListener('click', function(e) {
            const th = e.target.closest('th');
            if (th && th.dataset.sort) {
                sortStatuses(th.dataset.sort);
            }
        });
    }

        const settingsForm = document.getElementById('settings-form');
    if (settingsForm) {
        forceRefreshSettings();
        settingsForm.addEventListener('submit', saveSettings);
    } else {
    }
});

function loadDomains() {
    fetch('/api/domains')
        .then(response => response.json())
        .then(domains => {
            updateDomainList(domains);
        })
        .catch(error => console.error('Error:', error));
}

function updateDomainList(domains) {
    const domainList = document.getElementById('domain-list').getElementsByTagName('tbody')[0];
    domainList.innerHTML = '';
    domains.forEach(domain => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${domain}</td>
            <td>
                <button class="btn btn-sm delete-domain" data-domain="${domain}">删除</button>
            </td>
        `;
        domainList.appendChild(row);
    });
}

function addDomain(domain) {
    fetch('/domains', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: `domain=${encodeURIComponent(domain)}`
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            loadDomains(); // 重新加载域名列表
            loadDomainStatuses(); // 重新加载状态列表
            document.getElementById('new-domain').value = '';
        } else {
            alert('添加域名失败: ' + data.error);
        }
    })
    .catch(error => console.error('Error:', error));
}

function deleteDomain(domain) {
    fetch(`/domains/${encodeURIComponent(domain)}`, { 
        method: 'DELETE' 
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            loadDomains(); // 重新加载域名列表
            loadDomainStatuses(); // 重新加载状态列表
        } else {
            alert('删除域名失败: ' + data.error);
        }
    })
    .catch(error => console.error('Error:', error));
}


function loadWhoisServers() {
    fetch('/api/whois-servers')
        .then(response => response.json())
        .then(servers => {
            updateWhoisServerList(servers);
        })
        .catch(error => {
            console.error('Error loading Whois servers:', error);
        });
}

function updateWhoisServerList(servers) {
    const serverList = document.getElementById('whois-server-list').getElementsByTagName('tbody')[0];
    if (!serverList) return;

    serverList.innerHTML = '';
    for (const [tld, server] of Object.entries(servers)) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${tld}</td>
            <td>${server}</td>
            <td>
                <button class="btn  btn-sm delete-whois-server" data-tld="${tld}">删除</button>
            </td>
        `;
        serverList.appendChild(row);
    }
}

function addWhoisServer(tld, server) {
    fetch('/whois-servers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: `tld=${encodeURIComponent(tld)}&server=${encodeURIComponent(server)}`
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            loadWhoisServers();
            document.getElementById('new-tld').value = '';
            document.getElementById('new-server').value = '';
        } else {
            alert('添加 Whois 服务器失败: ' + data.error);
        }
    })
    .catch(error => console.error('Error:', error));
}

function deleteWhoisServer(tld) {
    fetch(`/whois-servers/${encodeURIComponent(tld)}`, { 
        method: 'DELETE' 
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            loadWhoisServers();
        } else {
            alert('删除 Whois 服务器失败: ' + data.error);
        }
    })
    .catch(error => console.error('Error:', error));
}

function loadDomainStatuses() {
    fetch('/domain-statuses')
        .then(response => response.json())
        .then(statuses => {
            updateDomainStatusList(statuses);
        })
        .catch(error => console.error('Error:', error));
}

function updateDomainStatusList(statuses) {
    const statusTableBody = document.getElementById('status-table-body');
    if (!statusTableBody) return;

    statusTableBody.innerHTML = '';
    statuses.forEach(status => {
        const row = document.createElement('tr');
        let registeredStatus, lastCheckedTime, monitorStatus;

        if (new Date(status.LastChecked).getFullYear() === 1) {
            // 处理新添加的域名
            registeredStatus = '未查询';
            lastCheckedTime = '/';
            monitorStatus = '等待监控';
        } else {
            registeredStatus = status.Registered ? '已注册' : '可注册';
            lastCheckedTime = new Date(status.LastChecked).toLocaleString();
            monitorStatus = status.CheckCount < 3 ? '正在监控' : '已通知';
        }

        row.innerHTML = `
            <td>${status.Domain}</td>
            <td>${registeredStatus}</td>
            <td>${lastCheckedTime}</td>
            <td>${monitorStatus}</td>
        `;
        statusTableBody.appendChild(row);
    });
}
function refreshDomainStatuses() {
    const button = document.getElementById('refresh-status-btn');
    button.disabled = true;
    button.textContent = '刷新中...';

    fetch('/refresh-statuses', { method: 'POST' })
        .then(response => response.json())
        .then(statuses => {
            updateDomainStatusList(statuses);
            button.disabled = false;
            button.textContent = '刷新状态';
        })
        .catch(error => {
            console.error('Error:', error);
            button.disabled = false;
            button.textContent = '刷新状态';
        });
}
document.addEventListener('DOMContentLoaded', function() {
    const themeController = document.querySelector('.theme-controller');
    
    // 检查本地存储中的主题设置
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
        document.documentElement.setAttribute('data-theme', savedTheme);
        themeController.checked = (savedTheme === 'dark');
    }

    // 监听主题切换
    themeController.addEventListener('change', function() {
        const newTheme = this.checked ? 'dark' : 'wireframe';
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    });
});

let currentSort = { column: null, direction: 'asc' };

function updateDomainStatusList(statuses) {
    const statusTableBody = document.getElementById('status-table-body');
    if (!statusTableBody) return;

    statusTableBody.innerHTML = '';
    statuses.forEach(status => {
        const row = document.createElement('tr');
        let statusText, lastCheckedTime, monitorStatus;

        if (new Date(status.LastChecked).getFullYear() === 1) {
            statusText = '未查询';
            lastCheckedTime = '/';
            monitorStatus = '等待监控';
        } else {
            // 使用 if-else 链来确定状态，确保只有一个状态被选中
            if (status.PendingDelete) {
                statusText = '<span class="text-red-500">待删除</span>';
            } else if (status.Redemption) {
                statusText = '<span class="text-orange-500">赎回期</span>';
            } else if (status.Registered) {
                statusText = '已注册';
            } else {
                statusText = '<span class="text-green-500">可注册</span>';
            }
            lastCheckedTime = new Date(status.LastChecked).toLocaleString();
            monitorStatus = status.CheckCount < 3 ? '正在监控' : '已通知';
        }

        row.innerHTML = `
            <td>${status.Domain}</td>
            <td>${statusText}</td>
            <td>${lastCheckedTime}</td>
            <td>${monitorStatus}</td>
        `;
        statusTableBody.appendChild(row);
    });
}

// 添加这个函数来定期刷新状态
function startStatusRefresh() {
    setInterval(() => {
        fetch('/domain-statuses')
            .then(response => response.json())
            .then(statuses => {
                updateDomainStatusList(statuses);
            })
            .catch(error => console.error('Error:', error));
    }, 30000); // 每30秒刷新一次
}

function sortStatuses(column) {
    const statusTable = document.getElementById('domain-status-list');
    const tbody = statusTable.querySelector('tbody');
    const rows = Array.from(tbody.querySelectorAll('tr'));

    if (currentSort.column === column) {
        currentSort.direction = currentSort.direction === 'asc' ? 'desc' : 'asc';
    } else {
        currentSort.column = column;
        currentSort.direction = 'asc';
    }

    rows.sort((a, b) => {
        let aValue = a.children[getColumnIndex(column)].textContent;
        let bValue = b.children[getColumnIndex(column)].textContent;

        if (column === 'lastChecked') {
            aValue = aValue === '/' ? new Date(0) : new Date(aValue);
            bValue = bValue === '/' ? new Date(0) : new Date(bValue);
        }

        if (aValue < bValue) return currentSort.direction === 'asc' ? -1 : 1;
        if (aValue > bValue) return currentSort.direction === 'asc' ? 1 : -1;
        return 0;
    });

    rows.forEach(row => tbody.appendChild(row));
    updateSortIndicators();
}

function getColumnIndex(column) {
    switch (column) {
        case 'status': return 1;
        case 'lastChecked': return 2;
        case 'monitorStatus': return 3;
        default: return 0;
    }
}

function updateSortIndicators() {
    const headers = document.querySelectorAll('#domain-status-list th[data-sort]');
    headers.forEach(header => {
        const arrow = currentSort.column === header.dataset.sort
            ? (currentSort.direction === 'asc' ? ' ↑' : ' ↓')
            : ' ↕';
        header.textContent = header.textContent.replace(/[↑↓↕]/, arrow);
    });
}

function loadSettings() {
    fetch('/api/settings')
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            updateSettingsForm(data);
        })
        .catch(error => {
            console.error('Error loading settings:', error);
            alert('加载设置时出错: ' + error.message);
        });
}
function saveSettings(e) {
    e.preventDefault();
    const form = document.getElementById('settings-form');
    const formData = new FormData(form);
    const settings = {};

    for (let [key, value] of formData.entries()) {
        settings[key] = value;
    }

    // 将端口和频率转换为数字
    if ('SMTP_PORT' in settings) settings.SMTP_PORT = parseInt(settings.SMTP_PORT, 10) || 0;
    if ('WEB_PORT' in settings) settings.WEB_PORT = parseInt(settings.WEB_PORT, 10) || 0;
    if ('QUERY_FREQUENCY_SECONDS' in settings) settings.QUERY_FREQUENCY_SECONDS = parseInt(settings.QUERY_FREQUENCY_SECONDS, 10) || 0;

    fetch('/api/settings', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(settings),
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        if (data.message) {
            alert(data.message);
            forceRefreshSettings();
        } else if (data.error) {
            throw new Error(data.error);
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('保存设置时出错: ' + error.message);
    });
}

function updateSettingsForm(data) {
    const form = document.getElementById('settings-form');
    if (!form) {
        console.error('Settings form not found');
        return;
    }

    const config = data.config || {};

    const fields = [
        'RECIPIENT_EMAIL', 'SMTP_SERVER', 'SMTP_PORT', 'SMTP_USERNAME', 'SMTP_PASSWORD',
        'WEB_PORT', 'AUTH_USERNAME', 'AUTH_PASSWORD', 'QUERY_FREQUENCY_SECONDS', 'SESSION_SECRET'
    ];

    fields.forEach(field => {
        const input = form.querySelector(`[name="${field}"]`);
        if (input && config[field] !== undefined) {
            input.value = config[field];
        }
    });
}
function forceRefreshSettings() {
    fetch('/api/settings', {
        method: 'GET',
        headers: {
            'Cache-Control': 'no-cache',
            'Pragma': 'no-cache'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        updateSettingsForm(data);
    })
    .catch(error => {
        console.error('Error loading settings:', error);
        alert('加载设置时出错: ' + error.message);
    });
}