# HetrixTools2ServerChan

一个用 Golang 构建的轻量级 webhook 服务，用于接收HetrixTools的状态变更通知，并通过 Server 酱推送。

## 快速开始

### Docker 部署

```bash

# 运行容器
docker run -d -p 8080:8080 \
  -e AUTH_TOKEN=your_secure_token \
  -e SERVER_CHAN_KEY=your_server_chan_key \
  -e PORT=8080 \
  -e TZ=Asia/Shanghai \
  --name webhook-container \
  ghcr.io/oxnme/hetrixtools2serverchan:latest
```

### 配置参数

| 环境变量 | 描述 | 默认值 |
|---------|------|--------|
| `PORT` | 服务监听端口 | `:8080` |
| `AUTH_TOKEN` | Bearer Token 认证密钥 | `default_auth_token_here` |
| `SERVER_CHAN_KEY` | Server 酱 API Key | `default_server_chan_key_here` |
| `TZ` | 时区 | `Asia/Shanghai` |

### Docker Compose

```shell 
curl -L https://raw.githubusercontent.com/oXnMe/HetrixTools2ServerChan/refs/heads/main/docker-compose.yml -o docker-compose.yml
docker compose up -d
```

## Server 酱集成

服务会自动将监控状态推送到 Server 酱，消息格式如下：

**标题:** `{监控名称}已{状态}`

**内容:**
```
{监控名称} [{监控分类}] {监控目标}已于{时间}离线/恢复

错误信息:
- 位置: 错误详情
- 位置: 错误详情
```
