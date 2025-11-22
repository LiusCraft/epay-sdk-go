// API 接口支付示例
// 适用于需要获取二维码或支付链接，由前端展示的场景
package main

import (
	"encoding/json"
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
func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>API 接口支付示例</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 15px 0; border-radius: 5px; }
        .method { display: inline-block; padding: 3px 8px; background: #1890ff; color: white; border-radius: 3px; font-size: 12px; }
        .method.get { background: #52c41a; }
        code { background: #f0f0f0; padding: 2px 5px; border-radius: 3px; }
        pre { background: #f5f5f5; padding: 15px; overflow-x: auto; }
        .test-form { margin-top: 30px; padding: 20px; background: #fafafa; border-radius: 5px; }
        input, select { padding: 8px; margin: 5px 0; }
        button { background: #1890ff; color: white; padding: 10px 20px; border: none; cursor: pointer; margin-top: 10px; }
        button:hover { background: #40a9ff; }
        #result { margin-top: 20px; padding: 15px; background: #f5f5f5; border-radius: 5px; display: none; }
    </style>
</head>
<body>
    <h1>API 接口支付示例</h1>

    <h2>接口列表</h2>

    <div class="endpoint">
        <p><span class="method">POST</span> <code>/api/payment/create</code></p>
        <p>创建支付订单，返回二维码或支付链接</p>
        <p>请求参数（JSON）：</p>
        <pre>{
    "pay_type": "alipay",  // 支付方式：alipay, wxpay, qqpay
    "amount": 0.01,        // 金额（元）
    "name": "商品名称"      // 商品名称
}</pre>
    </div>

    <div class="endpoint">
        <p><span class="method get">GET</span> <code>/api/payment/query?out_trade_no=xxx</code></p>
        <p>查询订单状态</p>
    </div>

    <div class="endpoint">
        <p><span class="method">POST</span> <code>/api/payment/notify</code></p>
        <p>支付回调通知（由 EPay 服务器调用）</p>
    </div>

    <div class="test-form">
        <h3>在线测试</h3>
        <form id="payForm">
            <div>
                <label>支付方式：</label>
                <select id="payType">
                    <option value="alipay">支付宝</option>
                    <option value="wxpay">微信支付</option>
                    <option value="qqpay">QQ钱包</option>
                </select>
            </div>
            <div>
                <label>商品名称：</label>
                <input type="text" id="name" value="测试商品">
            </div>
            <div>
                <label>金额（元）：</label>
                <input type="number" id="amount" value="0.01" step="0.01" min="0.01">
            </div>
            <button type="submit">创建支付订单</button>
        </form>

        <div id="result">
            <h4>响应结果：</h4>
            <pre id="resultContent"></pre>
            <div id="qrcode" style="margin-top: 15px;"></div>
        </div>
    </div>

    <script>
        document.getElementById('payForm').addEventListener('submit', async function(e) {
            e.preventDefault();

            const data = {
                pay_type: document.getElementById('payType').value,
                name: document.getElementById('name').value,
                amount: parseFloat(document.getElementById('amount').value)
            };

            try {
                const resp = await fetch('/api/payment/create', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                const result = await resp.json();

                document.getElementById('result').style.display = 'block';
                document.getElementById('resultContent').textContent = JSON.stringify(result, null, 2);

                // 显示二维码链接
                if (result.success && result.data) {
                    let qrHtml = '';
                    if (result.data.qr_code) {
                        qrHtml += '<p>二维码链接：<a href="' + result.data.qr_code + '" target="_blank">' + result.data.qr_code + '</a></p>';
                    }
                    if (result.data.pay_url) {
                        qrHtml += '<p>支付链接：<a href="' + result.data.pay_url + '" target="_blank">点击支付</a></p>';
                    }
                    document.getElementById('qrcode').innerHTML = qrHtml;
                } else {
                    document.getElementById('qrcode').innerHTML = '';
                }
            } catch (err) {
                document.getElementById('result').style.display = 'block';
                document.getElementById('resultContent').textContent = 'Error: ' + err.message;
            }
        });
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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
		NotifyURL:  "https://yourdomain.com/api/payment/notify", // 替换为你的回调地址
		ReturnURL:  "https://yourdomain.com/payment/success",    // 替换为你的跳转地址
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API 支付示例服务启动: http://localhost:%s", port)
	log.Printf("请设置环境变量: EPAY_PID, EPAY_KEY, EPAY_API_URL")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
