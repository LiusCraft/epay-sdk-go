// API 接口支付示例
// 适用于需要获取二维码或支付链接，由前端展示的场景
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	epay "github.com/liuscraft/epay-sdk-go"
)

// Order 内存中存储的订单信息
type Order struct {
	OutTradeNo string    `json:"out_trade_no"`
	TradeNo    string    `json:"trade_no"`
	PayType    string    `json:"pay_type"`
	Name       string    `json:"name"`
	Money      float64   `json:"money"`
	Status     int       `json:"status"` // 0=未支付, 1=已支付
	PayURL     string    `json:"pay_url"`
	QRCode     string    `json:"qr_code"`
	CreateTime time.Time `json:"create_time"`
	PayTime    time.Time `json:"pay_time,omitempty"`
}

// 内存订单存储
var (
	orders     = make(map[string]*Order)
	ordersLock sync.RWMutex
)

//go:embed templates/*.html
var templatesFS embed.FS

var (
	client    *epay.Client
	notifyURL string
	returnURL string
)

func init() {
	// 从环境变量读取配置
	pid, _ := strconv.Atoi(os.Getenv("EPAY_PID"))
	if pid == 0 {
		pid = 1001 // 默认值，仅用于演示
	}

	key := os.Getenv("EPAY_KEY")
	if key == "" {
		key = "your-merchant-key" // 默认值，仅用于演示
	}

	apiURL := os.Getenv("EPAY_API_URL")
	if apiURL == "" {
		apiURL = "https://pay.example.com" // 默认值，仅用于演示
	}

	notifyURL = os.Getenv("EPAY_NOTIFY_URL")
	if notifyURL == "" {
		notifyURL = "http://localhost:8080/api/payment/notify" // 默认值，仅用于演示
	}

	returnURL = os.Getenv("EPAY_RETURN_URL")
	if returnURL == "" {
		returnURL = "http://localhost:8080/payment/return" // 默认值，仅用于演示
	}

	var err error
	client, err = epay.NewClient(&epay.Config{
		PID:        pid,
		Key:        key,
		APIBaseURL: apiURL,
		Timeout:    30,
		Debug:      true,
	})
	if err != nil {
		log.Fatalf("Failed to create epay client: %v", err)
	}
}

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func jsonResponse(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// indexHandler 首页 - 显示 API 文档
// HTML 文件已通过 embed 嵌入到二进制文件中，无需外部依赖
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 从嵌入的文件系统读取 HTML
	htmlContent, err := templatesFS.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "无法加载页面", http.StatusInternalServerError)
		log.Printf("读取 index.html 失败: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}

// createPaymentHandler 创建支付订单 API
func createPaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// 解析请求
	var req struct {
		PayType string  `json:"pay_type"`
		Amount  float64 `json:"amount"`
		Name    string  `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// 参数验证
	if req.Amount <= 0 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Amount must be greater than 0",
		})
		return
	}

	if req.Name == "" {
		req.Name = "商品"
	}

	if req.PayType == "" {
		req.PayType = "alipay"
	}

	// 生成唯一订单号
	outTradeNo := fmt.Sprintf("API%d", time.Now().UnixNano())

	// 获取客户端 IP
	clientIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientIP = xff
	}

	// 调用 SDK 创建支付
	resp, err := client.CreatePayment(&epay.PaymentRequest{
		Type:       req.PayType,
		OutTradeNo: outTradeNo,
		NotifyURL:  notifyURL,
		ReturnURL:  returnURL,
		Name:       req.Name,
		Money:      req.Amount,
		ClientIP:   clientIP,
		Device:     "pc",
		Param:      `{"source": "api_example"}`, // 可选：业务扩展参数
	})

	if err != nil {
		log.Printf("创建支付订单失败: %v", err)
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Message: "Create payment failed: " + err.Error(),
		})
		return
	}

	log.Printf("创建API支付订单: %s, 金额: %.2f", outTradeNo, req.Amount)

	// 存储订单到内存
	ordersLock.Lock()
	orders[outTradeNo] = &Order{
		OutTradeNo: outTradeNo,
		TradeNo:    resp.TradeNo,
		PayType:    req.PayType,
		Name:       req.Name,
		Money:      req.Amount,
		Status:     0,
		PayURL:     resp.PayURL,
		QRCode:     resp.QRCode,
		CreateTime: time.Now(),
	}
	ordersLock.Unlock()

	// 返回支付信息
	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"out_trade_no": outTradeNo,
			"trade_no":     resp.TradeNo,
			"pay_url":      resp.PayURL,
			"qr_code":      resp.QRCode,
			"url_scheme":   resp.URLScheme,
		},
	})
}

// listOrdersHandler 获取所有订单列表
func listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ordersLock.RLock()
	orderList := make([]*Order, 0, len(orders))
	for _, order := range orders {
		orderList = append(orderList, order)
	}
	ordersLock.RUnlock()

	// 按创建时间倒序排列
	sort.Slice(orderList, func(i, j int) bool {
		return orderList[i].CreateTime.After(orderList[j].CreateTime)
	})

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    orderList,
	})
}

// queryOrderHandler 查询订单状态 API
func queryOrderHandler(w http.ResponseWriter, r *http.Request) {
	outTradeNo := r.URL.Query().Get("out_trade_no")
	tradeNo := r.URL.Query().Get("trade_no")

	if outTradeNo == "" && tradeNo == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "out_trade_no or trade_no is required",
		})
		return
	}

	// 调用 SDK 查询订单
	order, err := client.QueryOrder(&epay.OrderQueryRequest{
		OutTradeNo: outTradeNo,
		TradeNo:    tradeNo,
	})

	if err != nil {
		log.Printf("查询订单失败: %v", err)
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Message: "Query order failed: " + err.Error(),
		})
		return
	}

	// 返回订单信息
	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"out_trade_no": order.OutTradeNo,
			"trade_no":     order.TradeNo,
			"type":         order.Type,
			"name":         order.Name,
			"money":        order.Money,
			"status":       order.Status, // 1=已支付, 0=未支付
			"add_time":     order.AddTime,
			"end_time":     order.EndTime,
		},
	})
}

// notifyHandler 支付回调处理
func notifyHandler(w http.ResponseWriter, r *http.Request) {
	// 解析回调参数
	params := epay.ParseNotifyParams(r)

	log.Printf("收到支付回调: %+v", params)

	// 验证签名
	notifyData, err := client.VerifyNotify(params)
	if err != nil {
		log.Printf("签名验证失败: %v", err)
		w.Write([]byte("fail"))
		return
	}

	// 检查支付状态
	if notifyData.TradeStatus == "TRADE_SUCCESS" {
		log.Printf("订单支付成功: %s, 金额: %s, EPay订单号: %s",
			notifyData.OutTradeNo, notifyData.Money, notifyData.TradeNo)

		// 更新内存中的订单状态
		ordersLock.Lock()
		if order, exists := orders[notifyData.OutTradeNo]; exists {
			order.Status = 1
			order.PayTime = time.Now()
			order.TradeNo = notifyData.TradeNo
		}
		ordersLock.Unlock()

		// TODO: 在这里处理你的业务逻辑
		// 1. 检查订单是否已处理（幂等性）
		// 2. 验证金额是否正确
		// 3. 更新订单状态
		// 4. 发放商品/服务

		// 解析业务扩展参数
		if notifyData.Param != "" {
			log.Printf("业务参数: %s", notifyData.Param)
		}
	}

	// 返回 success 告知 EPay 已收到通知
	w.Write([]byte("success"))
}

// returnHandler 同步回调页面
func returnHandler(w http.ResponseWriter, r *http.Request) {
	params := epay.ParseNotifyParams(r)
	outTradeNo := params["out_trade_no"]

	// 验证签名
	_, err := client.VerifyNotify(params)
	if err != nil {
		log.Printf("Return 签名验证失败: %v", err)
		http.Error(w, "签名验证失败", http.StatusBadRequest)
		return
	}

	// 从内存获取订单信息
	ordersLock.RLock()
	order, exists := orders[outTradeNo]
	ordersLock.RUnlock()

	var orderData map[string]interface{}
	if exists {
		orderData = map[string]interface{}{
			"out_trade_no": order.OutTradeNo,
			"trade_no":     order.TradeNo,
			"pay_type":     order.PayType,
			"name":         order.Name,
			"money":        order.Money,
			"status":       order.Status,
			"create_time":  order.CreateTime.Format("2006-01-02 15:04:05"),
		}
	} else {
		// 如果内存中没有，从回调参数获取
		orderData = map[string]interface{}{
			"out_trade_no": params["out_trade_no"],
			"trade_no":     params["trade_no"],
			"pay_type":     params["type"],
			"name":         params["name"],
			"money":        params["money"],
			"status":       1, // 回调到 return 通常表示支付成功
		}
	}

	// 返回 JSON（前端可以根据需要渲染）
	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "支付完成",
		Data:    orderData,
	})
}

// refundHandler 退款申请 API
func refundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// 解析请求
	var req struct {
		OutTradeNo string  `json:"out_trade_no"`
		TradeNo    string  `json:"trade_no"`
		Money      float64 `json:"money"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.OutTradeNo == "" && req.TradeNo == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "out_trade_no or trade_no is required",
		})
		return
	}

	if req.Money <= 0 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "money must be greater than 0",
		})
		return
	}

	// 调用 SDK 申请退款
	resp, err := client.Refund(&epay.RefundRequest{
		OutTradeNo: req.OutTradeNo,
		TradeNo:    req.TradeNo,
		Money:      req.Money,
	})

	if err != nil {
		log.Printf("退款申请失败: %v", err)
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Message: "Refund failed: " + err.Error(),
		})
		return
	}

	if resp.Code != 1 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: resp.Msg,
		})
		return
	}

	log.Printf("退款申请成功: %s, 金额: %.2f", req.OutTradeNo, req.Money)

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Refund submitted successfully",
	})
}

func main() {
	// 注册路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/payment/create", createPaymentHandler)
	http.HandleFunc("/api/payment/query", queryOrderHandler)
	http.HandleFunc("/api/payment/notify", notifyHandler)
	http.HandleFunc("/api/payment/refund", refundHandler)
	http.HandleFunc("/api/orders", listOrdersHandler)
	http.HandleFunc("/payment/return", returnHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API 支付示例服务启动: http://localhost:%s", port)
	log.Printf("请设置环境变量: EPAY_PID, EPAY_KEY, EPAY_API_URL, EPAY_NOTIFY_URL, EPAY_RETURN_URL")
	log.Printf("当前配置: NotifyURL=%s, ReturnURL=%s", notifyURL, returnURL)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
