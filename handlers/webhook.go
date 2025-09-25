package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"webhook/types"
)

type MonitorData struct {
	MonitorID       string            `json:"monitor_id"`
	MonitorName     string            `json:"monitor_name"`
	MonitorTarget   string            `json:"monitor_target"`
	MonitorType     string            `json:"monitor_type"`
	MonitorCategory string            `json:"monitor_category"`
	MonitorStatus   string            `json:"monitor_status"`
	Timestamp       int64             `json:"timestamp"`
	MonitorErrors   map[string]string `json:"monitor_errors,omitempty"`
}

type ServerChanRequest struct {
	Title string `json:"title"`
	Desp  string `json:"desp"`
}

type ServerChanResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		PushID    string `json:"pushid"`
		ReadKey   string `json:"readkey"`
		Error     string `json:"error"`
		ErrorCode int    `json:"errorcode"`
	} `json:"data"`
}

// UnixToDateTime 将UNIX时间戳转换为指定时区的格式化日期时间字符串
func UnixToDateTime(timestamp int64, location *time.Location) string {
	t := time.Unix(timestamp, 0).In(location)
	return t.Format("2006-01-02 15:04:05")
}

func WebhookHandler(cfg *types.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 验证Authorization头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		expectedToken := "Bearer " + cfg.Server.AuthToken
		if authHeader != expectedToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 读取请求体
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// 打印接收到的原始请求体
		log.Printf("Received webhook data: %s", string(body))

		// 解析JSON
		var data MonitorData
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// 验证必需字段
		if data.MonitorName == "" || data.MonitorTarget == "" || data.MonitorStatus == "" || data.Timestamp == 0 {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// 打印解析后的监控数据
		convertedTime := UnixToDateTime(data.Timestamp, cfg.Server.TimeLocation)
		log.Printf("Parsed monitor data: Name=%s, Target=%s, Status=%s, Type=%s",
			data.MonitorName, data.MonitorTarget, data.MonitorStatus, data.MonitorType)
		log.Printf("Timestamp: %d -> %s (Timezone: %s)",
			data.Timestamp, convertedTime, cfg.Server.TimeLocation.String())

		// 处理数据并发送到Server酱
		err = sendToServerChan(cfg, data)
		if err != nil {
			log.Printf("Error sending to ServerChan: %v", err)
			http.Error(w, "Error sending to ServerChan", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Webhook processed successfully"))
	}
}

func sendToServerChan(cfg *types.Config, data MonitorData) error {
	// 构建标题 - 使用 monitor_name
	statusText := "恢复"
	if data.MonitorStatus == "offline" {
		statusText = "离线"
	}
	title := fmt.Sprintf("%s已%s", data.MonitorName, statusText)

	// 限制标题长度
	if len(title) > 32 {
		title = title[:32]
	}

	// 构建描述 - 使用指定时区的时间
	datetime := UnixToDateTime(data.Timestamp, cfg.Server.TimeLocation)
	desp := fmt.Sprintf("%s %s %s已于%s%s",
		data.MonitorName,
		data.MonitorCategory,
		data.MonitorTarget,
		datetime,
		statusText)

	// 如果是离线状态，添加错误信息
	if data.MonitorStatus == "offline" && len(data.MonitorErrors) > 0 {
		desp += "\n\n**错误信息:**"
		for location, errorMsg := range data.MonitorErrors {
			desp += fmt.Sprintf("\n- %s: %s", location, errorMsg)
		}
	}

	// 构建Server酱请求
	serverChanReq := ServerChanRequest{
		Title: title,
		Desp:  desp,
	}

	// 转换为JSON
	jsonData, err := json.Marshal(serverChanReq)
	if err != nil {
		return err
	}

	// 打印将要发送到Server酱的数据
	log.Printf("Using timezone: %s", cfg.Server.TimeLocation.String())
	log.Printf("Sending to ServerChan: %s", string(jsonData))
	log.Printf("ServerChan URL: https://sctapi.ftqq.com/%s.send", cfg.ServerChan.APIKey)

	// 发送到Server酱
	url := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", cfg.ServerChan.APIKey)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取Server酱的响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response body: %v", err)
	}

	// 打印Server酱的原始响应
	log.Printf("ServerChan response status: %s", resp.Status)
	log.Printf("ServerChan response body: %s", string(respBody))

	// 解析Server酱的响应JSON
	var serverChanResp ServerChanResponse
	if err := json.Unmarshal(respBody, &serverChanResp); err != nil {
		log.Printf("Warning: Failed to parse ServerChan JSON response: %v", err)
	} else {
		// 打印结构化的响应信息
		if serverChanResp.Code == 0 {
			log.Printf("ServerChan push successful! PushID: %s", serverChanResp.Data.PushID)
		} else {
			log.Printf("ServerChan push failed! Code: %d, Message: %s, Error: %s",
				serverChanResp.Code, serverChanResp.Message, serverChanResp.Data.Error)
		}
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ServerChan API returned non-200 status: %s, body: %s", resp.Status, string(respBody))
	}

	return nil
}
