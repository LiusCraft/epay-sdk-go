// Form 表单支付示例
// 适用于网页端用户直接跳转到支付页面的场景
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	epay "github.com/liuscraft/epay-sdk-go"
)

var client *epay.Client

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
func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Form 表单支付示例</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, select { width: 100%; padding: 8px; box-sizing: border-box; }
        button { background: #1890ff; color: white; padding: 10px 20px; border: none; cursor: pointer; }
        button:hover { background: #40a9ff; }
        h1 { color: #333; }
        .methods { margin: 20px 0; }
        .methods a { display: inline-block; margin-right: 10px; padding: 10px 15px; background: #f0f0f0; text-decoration: none; color: #333; }
        .methods a:hover { background: #e0e0e0; }
    </style>
</head>
<body>
    <h1>Form 表单支付示例</h1>

    <div class="methods">
        <h3>选择支付方式：</h3>
        <a href="/pay/form?type=alipay&money=0.01&name=测试商品">支付宝支付</a>
        <a href="/pay/form?type=wxpay&money=0.01&name=测试商品">微信支付</a>
        <a href="/pay/form?type=&money=0.01&name=测试商品">收银台（选择支付方式）</a>
    </div>

    <h3>或者自定义参数：</h3>
    <form action="/pay/form" method="GET">
        <div class="form-group">
            <label>支付方式</label>
            <select name="type">
                <option value="">收银台（用户选择）</option>
                <option value="alipay">支付宝</option>
                <option value="wxpay">微信支付</option>
                <option value="qqpay">QQ钱包</option>
            </select>
        </div>
        <div class="form-group">
            <label>商品名称</label>
            <input type="text" name="name" value="测试商品" required>
        </div>
        <div class="form-group">
            <label>金额（元）</label>
            <input type="number" name="money" value="0.01" step="0.01" min="0.01" required>
        </div>
        <button type="submit">发起支付</button>
    </form>

    <hr style="margin-top: 30px;">
    <p><a href="/pay/url?type=alipay&money=0.01&name=测试商品">使用 URL 跳转方式支付</a></p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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
		NotifyURL:  "https://yourdomain.com/api/payment/notify", // 替换为你的回调地址
		ReturnURL:  "https://yourdomain.com/payment/success",    // 替换为你的跳转地址
		Name:       name,
		Money:      money,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("创建支付失败: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("创建表单支付订单: %s, 金额: %.2f", outTradeNo, money)

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
		NotifyURL:  "https://yourdomain.com/api/payment/notify",
		ReturnURL:  "https://yourdomain.com/payment/success",
		Name:       name,
		Money:      money,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("创建支付失败: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("创建URL支付订单: %s, 金额: %.2f, URL: %s", outTradeNo, money, payURL)

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

		// TODO: 在这里处理你的业务逻辑
		// 1. 检查订单是否已处理（幂等性）
		// 2. 验证金额是否正确
		// 3. 更新订单状态
		// 4. 发放商品/服务
	}

	// 返回 success 告知 EPay 已收到通知
	w.Write([]byte("success"))
}

// returnHandler 支付成功跳转页面
func returnHandler(w http.ResponseWriter, r *http.Request) {
	// 同步跳转，可以验证签名
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
}

func main() {
	// 注册路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/pay/form", formPaymentHandler)
	http.HandleFunc("/pay/url", urlPaymentHandler)
	http.HandleFunc("/api/payment/notify", notifyHandler)
	http.HandleFunc("/payment/success", returnHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Form 支付示例服务启动: http://localhost:%s", port)
	log.Printf("请设置环境变量: EPAY_PID, EPAY_KEY, EPAY_API_URL")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
