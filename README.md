<div align="center">
    <h1>Puff</h1>
<div align="center">
    
![GitHub repo size](https://img.shields.io/github/repo-size/bitaur/puff)
![GitHub Repo stars](https://img.shields.io/github/stars/bitaur/puff)
![GitHub all releases](https://img.shields.io/github/downloads/bitaur/puff/total)

</div>
    <p>开源、快速、便捷、基于Go的域名监控程序。</p>
</div>
    <p>原理：通过 Whois 通过字段进行判断域名状态。</p>

<img src="https://s2.loli.net/2024/10/01/9sfwrgutUqOaAxd.png"/>

## 功能

- [x] 域名赎回期、可注册、待删除状态通知
- [x] 邮箱通知
- [ ] Telegarm通知
- [ ] 域名抢注

# 部署 Puff

目前支持三种部署方式，编译部署、手动部署、Docker 部署。

## 编译部署

### 环境要求

- Go 版本 >=1.22.0

### 克隆仓库

``` shell
git clone https://github.com/bitaur/puff.git
```

### 构建程序

``` shell
go build -o Puff.exe main.go
```

### 运行

```
./Puff
```



## 手动部署

### 下载 Puff

打开 [Puff Release](https://github.com/BitAUR/Puff/releases) 下载对应的平台以及系统的文件。

如果最新的包没有您对应的二进制文件，可以提交 [issues](https://github.com/BitAUR/Puff/issues) ，或可以选择自己编译安装。

其中：

armv6 对应 arm 架构32位版本，arm64 对应 arm 架构64位版本。

x86 对应 x86 平台32位版本，x86_64 对应  x86 平台64位版本。

克隆模板文件以及静态资源文件。

### 手动运行

#### Linux / MacOS

``` shell
# 解压下载后的文件，请求改为您下载的文件名
tar -zxvf filename.tar.gz

# 授予执行权限
chmod +x Puff

./Puff
```

#### Windows

双击运行即可。

### 持久化运行

#### Linux

使用编辑器编辑 ``` /usr/lib/systemd/system/puff.service``` 添加如下内容：

``` shell
[Unit]
Description=puff
After=network.target
 
[Service]
Type=simple
WorkingDirectory=puff_path
ExecStart=puff_path/Puff
Restart=on-failure
 
[Install]
WantedBy=multi-user.target
```

保存后，使用 ```systemctl deamon-reload``` 重载配置。具体使用命令如下：

- 启动: `systemctl start puff`
- 关闭: `systemctl stop puff`
- 配置开机自启: `systemctl enable puff`
- 取消开机自启: `systemctl disable puff`
- 状态: `systemctl status puff`
- 重启: `systemctl restart puff`

### 更新版本

如果有新版本更新，下载新版本，将旧版本的文件删除即可。

## Docker 部署

首先请确保您正确的安装并配置了 Docker 以及 Docker Compose

### Docker CLI

``` shell
docker run -d --restart=unless-stopped -v /data/puff:/data -p 8080:8080 --name="Puff" bitaur/puff:latest
```

### Docker Compose

在空目录中创建 docker-compose.yaml 文件，将下列内容保存。

``` dockerfile
services:
  Puff:
    image: bitaur/puff:latest
    container_name: Puff
    volumes:
      - /data/puff:/data
    restart: unless-stopped
    ports:
      - 8080:8080
```

保存后，使用 ``` docker compose up -d``` 创建并启动容器。

### Docker 容器更新

#### CLI

```shell
#查看容器ID
docker ps -a

#停止容器
docker stop ID

#删除容器
docker rm ID

#获取新镜像
docker pull bitaur/puff:latest

# 输入安装命令
docker run -d --restart=unless-stopped -v /data/puff:/data -p 9740:9740 --name="Puff" bitaur/puff:latest
```

#### Docker Compose

``` shell
#获取新镜像
docker pull bitaur/puff:latest

#创建并启动容器
docker compose up -d
```

## 访问

此时打开 `localhost:8080` 即可打开站点。默认账号密码均为 `admin`。
