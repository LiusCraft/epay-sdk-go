// Form 表单支付示例
// 适用于网页端用户直接跳转到支付页面的场景
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
		returnURL = "http://localhost:8080/payment/success" // 默认值，仅用于演示
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

// indexHandler 首页 - 显示支付表单
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

// formPaymentHandler 生成 HTML 表单跳转支付
func formPaymentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	payType := r.URL.Query().Get("type")
	name := r.URL.Query().Get("name")
	moneyStr := r.URL.Query().Get("money")

	if name == "" {
		name = "测试商品"
	}

	money := 0.01
	if moneyStr != "" {
		if m, err := strconv.ParseFloat(moneyStr, 64); err == nil {
			money = m
		}
	}

	// 生成唯一订单号
	outTradeNo := fmt.Sprintf("FORM%d", time.Now().UnixNano())

	// 构建表单支付请求
	htmlForm, err := client.BuildFormPayment(&epay.FormPaymentRequest{
		Type:       payType, // 为空则显示收银台
		OutTradeNo: outTradeNo,
		NotifyURL:  notifyURL,
		ReturnURL:  returnURL,
		Name:       name,
		Money:      money,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("创建支付失败: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("创建表单支付订单: %s, 金额: %.2f", outTradeNo, money)

	// 存储订单到内存
	ordersLock.Lock()
	orders[outTradeNo] = &Order{
		OutTradeNo: outTradeNo,
		PayType:    payType,
		Name:       name,
		Money:      money,
		Status:     0,
		CreateTime: time.Now(),
	}
	ordersLock.Unlock()

	// 返回 HTML 表单，浏览器会自动提交跳转到支付页面
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlForm))
}

// urlPaymentHandler 生成 URL 跳转支付
func urlPaymentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	payType := r.URL.Query().Get("type")
	name := r.URL.Query().Get("name")
	moneyStr := r.URL.Query().Get("money")

	if name == "" {
		name = "测试商品"
	}

	money := 0.01
	if moneyStr != "" {
		if m, err := strconv.ParseFloat(moneyStr, 64); err == nil {
			money = m
		}
	}

	// 生成唯一订单号
	outTradeNo := fmt.Sprintf("URL%d", time.Now().UnixNano())

	// 构建支付 URL
	payURL, err := client.BuildFormPaymentURL(&epay.FormPaymentRequest{
		Type:       payType,
		OutTradeNo: outTradeNo,
		NotifyURL:  notifyURL,
		ReturnURL:  returnURL,
		Name:       name,
		Money:      money,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("创建支付失败: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("创建URL支付订单: %s, 金额: %.2f, URL: %s", outTradeNo, money, payURL)

	// 存储订单到内存
	ordersLock.Lock()
	orders[outTradeNo] = &Order{
		OutTradeNo: outTradeNo,
		PayType:    payType,
		Name:       name,
		Money:      money,
		Status:     0,
		CreateTime: time.Now(),
	}
	ordersLock.Unlock()

	// 直接跳转到支付页面
	http.Redirect(w, r, payURL, http.StatusFound)
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
		log.Printf("订单支付成功: %s, 金额: %s", notifyData.OutTradeNo, notifyData.Money)

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
	}

	// 返回 success 告知 EPay 已收到通知
	w.Write([]byte("success"))
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    orderList,
	})
}

// returnHandler 支付成功跳转页面
// HTML 模板已通过 embed 嵌入，使用 %s 占位符，分别对应订单号和金额
func returnHandler(w http.ResponseWriter, r *http.Request) {
	// 同步跳转，可以验证签名
	params := epay.ParseNotifyParams(r)

	// 从嵌入的文件系统读取 HTML 模板
	htmlTemplate, err := templatesFS.ReadFile("templates/return.html")
	if err != nil {
		http.Error(w, "无法加载页面", http.StatusInternalServerError)
		log.Printf("读取 return.html 失败: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, string(htmlTemplate), params["out_trade_no"], params["money"])
}

func main() {
	// 注册路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/pay/form", formPaymentHandler)
	http.HandleFunc("/pay/url", urlPaymentHandler)
	http.HandleFunc("/api/payment/notify", notifyHandler)
	http.HandleFunc("/api/orders", listOrdersHandler)
	http.HandleFunc("/payment/success", returnHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Form 支付示例服务启动: http://localhost:%s", port)
	log.Printf("请设置环境变量: EPAY_PID, EPAY_KEY, EPAY_API_URL, EPAY_NOTIFY_URL, EPAY_RETURN_URL")
	log.Printf("当前配置: NotifyURL=%s, ReturnURL=%s", notifyURL, returnURL)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
