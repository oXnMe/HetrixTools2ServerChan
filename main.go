package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"webhook/handlers"
	"webhook/types"
)

const (
	defaultPort      = ":8080"
	defaultAuthToken = "default_auth_token_here"
	defaultAPIKey    = "default_server_chan_key_here"
	defaultTimeZone  = "Asia/Shanghai"
)

func main() {
	// 加载配置（从环境变量读取，未配置则使用默认值）
	cfg := loadConfig()

	// 设置路由
	http.HandleFunc("/webhook", handlers.WebhookHandler(cfg))

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Auth token: %s", maskToken(cfg.Server.AuthToken))
	log.Printf("ServerChan API key: %s", maskToken(cfg.ServerChan.APIKey))
	log.Printf("Time zone: %s", cfg.Server.TimeLocation.String())

	log.Fatal(http.ListenAndServe(cfg.Server.Port, nil))
}

func loadConfig() *types.Config {
	cfg := &types.Config{}

	// 从环境变量读取配置，未设置则使用默认值
	cfg.Server.Port = normalizePort(getEnv("PORT", defaultPort))
	cfg.Server.AuthToken = getEnv("AUTH_TOKEN", defaultAuthToken)
	timeZone := getEnv("TZ", defaultTimeZone)
	cfg.ServerChan.APIKey = getEnv("SERVER_CHAN_KEY", defaultAPIKey)

	// 设置时区位置
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		log.Printf("Warning: Invalid timezone '%s', using UTC. Error: %v", timeZone, err)
		location = time.UTC
	}
	cfg.Server.TimeLocation = location
	cfg.Server.TimeZone = timeZone

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func normalizePort(port string) string {
	if port == "" {
		return defaultPort
	}

	if len(port) > 0 && port[0] == ':' {
		return port
	}

	if _, err := strconv.Atoi(port); err == nil {
		return ":" + port
	}

	log.Printf("Warning: Invalid port format '%s', using default port %s", port, defaultPort)
	return defaultPort
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}
