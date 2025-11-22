// Package handler 提供标准的 http.Handler 实现
// 可以直接集成到任何 Go Web 框架中（net/http, Gin, Echo, Chi 等）
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	epay "github.com/liuscraft/epay-sdk-go"
)

// NotifyCallback 支付回调处理函数
// 参数: notifyData - 验证签名后的回调数据
// 返回: error - 如果返回 error，会向 EPay 返回 "fail"
type NotifyCallback func(notifyData *epay.NotifyData) error

// Handlers EPay HTTP 处理器集合
type Handlers struct {
	client    *epay.Client
	notifyURL string
	returnURL string
	logger    Logger
}

// Logger 日志接口
type Logger interface {
	Printf(format string, v ...interface{})
}

// Option 配置选项
type Option func(*Handlers)

// WithNotifyURL 设置异步回调地址
func WithNotifyURL(url string) Option {
	return func(h *Handlers) {
		h.notifyURL = url
	}
}

// WithReturnURL 设置同步跳转地址
func WithReturnURL(url string) Option {
	return func(h *Handlers) {
		h.returnURL = url
	}
}

// WithLogger 设置自定义日志器
func WithLogger(logger Logger) Option {
	return func(h *Handlers) {
		h.logger = logger
	}
}

// NewHandlers 创建 HTTP 处理器集合
// 使用示例:
//
//	handlers := handler.NewHandlers(client,
//	    handler.WithNotifyURL("https://yourdomain.com/notify"),
//	    handler.WithReturnURL("https://yourdomain.com/return"),
//	)
//	http.Handle("/pay/form", handlers.FormPayment())
//	http.Handle("/notify", handlers.Notify(callback))
func NewHandlers(client *epay.Client, opts ...Option) *Handlers {
	h := &Handlers{
		client: client,
		logger: log.Default(),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// FormPayment 返回表单支付 Handler
// 从 URL 参数读取：type, name, money
// 生成 HTML 表单自动提交到支付页面
func (h *Handlers) FormPayment() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析参数
		payType := r.URL.Query().Get("type")
		name := r.URL.Query().Get("name")
		moneyStr := r.URL.Query().Get("money")

		if name == "" {
			name = "商品"
		}

		money, err := strconv.ParseFloat(moneyStr, 64)
		if err != nil || money <= 0 {
			http.Error(w, "Invalid money parameter", http.StatusBadRequest)
			return
		}

		// 生成订单号
		outTradeNo := fmt.Sprintf("ORDER%d", time.Now().UnixNano())

		// 构建表单
		htmlForm, err := h.client.BuildFormPayment(&epay.FormPaymentRequest{
			Type:       payType,
			OutTradeNo: outTradeNo,
			NotifyURL:  h.notifyURL,
			ReturnURL:  h.returnURL,
			Name:       name,
			Money:      money,
		})

		if err != nil {
			h.logger.Printf("Build form payment failed: %v", err)
			http.Error(w, "Failed to create payment", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlForm))
	})
}

// QRCodePayment 返回二维码支付 Handler（API 模式）
// 请求体为 JSON: {"pay_type": "alipay", "name": "商品", "money": 0.01}
// 返回支付信息 JSON
func (h *Handlers) QRCodePayment() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 解析请求
		var req struct {
			PayType string  `json:"pay_type"`
			Name    string  `json:"name"`
			Money   float64 `json:"money"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid request body",
			})
			return
		}

		if req.Money <= 0 {
			h.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid money",
			})
			return
		}

		if req.Name == "" {
			req.Name = "商品"
		}

		// 生成订单号
		outTradeNo := fmt.Sprintf("API%d", time.Now().UnixNano())

		// 获取客户端 IP
		clientIP := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			clientIP = xff
		}

		// 创建支付
		resp, err := h.client.CreatePayment(&epay.PaymentRequest{
			Type:       req.PayType,
			OutTradeNo: outTradeNo,
			NotifyURL:  h.notifyURL,
			ReturnURL:  h.returnURL,
			Name:       req.Name,
			Money:      req.Money,
			ClientIP:   clientIP,
			Device:     "pc",
		})

		if err != nil {
			h.logger.Printf("Create payment failed: %v", err)
			h.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to create payment",
			})
			return
		}

		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"out_trade_no": outTradeNo,
				"trade_no":     resp.TradeNo,
				"pay_url":      resp.PayURL,
				"qr_code":      resp.QRCode,
				"url_scheme":   resp.URLScheme,
			},
		})
	})
}

// Notify 返回支付回调 Handler
// callback 函数用于处理业务逻辑，如果返回 error，会向 EPay 返回 "fail"
func (h *Handlers) Notify(callback NotifyCallback) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析回调参数
		params := epay.ParseNotifyParams(r)

		h.logger.Printf("Received payment notify: %+v", params)

		// 验证签名
		notifyData, err := h.client.VerifyNotify(params)
		if err != nil {
			h.logger.Printf("Verify notify signature failed: %v", err)
			w.Write([]byte("fail"))
			return
		}

		// 执行业务回调
		if callback != nil {
			if err := callback(notifyData); err != nil {
				h.logger.Printf("Notify callback failed: %v", err)
				w.Write([]byte("fail"))
				return
			}
		}

		// 返回成功
		w.Write([]byte("success"))
	})
}

// Return 返回支付同步跳转 Handler
// 返回简单的成功页面，可以自定义
func (h *Handlers) Return() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := epay.ParseNotifyParams(r)

		html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>支付结果</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; }
        .success { color: #52c41a; font-size: 24px; }
        .info { margin: 20px 0; color: #666; }
    </style>
</head>
<body>
    <h1 class="success">支付完成</h1>
    <div class="info">
        <p>订单号: %s</p>
        <p>支付金额: %s 元</p>
    </div>
    <p><a href="/">返回首页</a></p>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, html, params["out_trade_no"], params["money"])
	})
}

// QueryOrder 返回订单查询 Handler
// URL 参数: out_trade_no 或 trade_no
func (h *Handlers) QueryOrder() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		outTradeNo := r.URL.Query().Get("out_trade_no")
		tradeNo := r.URL.Query().Get("trade_no")

		if outTradeNo == "" && tradeNo == "" {
			h.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "out_trade_no or trade_no is required",
			})
			return
		}

		order, err := h.client.QueryOrder(&epay.OrderQueryRequest{
			OutTradeNo: outTradeNo,
			TradeNo:    tradeNo,
		})

		if err != nil {
			h.logger.Printf("Query order failed: %v", err)
			h.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Query failed",
			})
			return
		}

		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    order,
		})
	})
}

// writeJSON 写入 JSON 响应
func (h *Handlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
