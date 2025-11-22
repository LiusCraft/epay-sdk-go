// Quick 快速启动示例
// 演示如何用几行代码快速集成 EPay 支付
package main

import (
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
	"github.com/liuscraft/epay-sdk-go/handler"
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

func main() {
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

	notifyURL := os.Getenv("EPAY_NOTIFY_URL")
	if notifyURL == "" {
		notifyURL = "http://localhost:8080/notify" // 默认值，仅用于演示
	}

	returnURL := os.Getenv("EPAY_RETURN_URL")
	if returnURL == "" {
		returnURL = "http://localhost:8080/return" // 默认值，仅用于演示
	}

	// 方式 1: 一行代码创建客户端（使用默认配置）
	client := epay.NewQuick(pid, key, apiURL)

	// 方式 2: 链式 API（更多配置选项）
	// client := epay.New(pid, key, apiURL).
	//     WithTimeout(30).
	//     WithDebug(true).
	//     MustBuild()

	// 创建 HTTP Handlers
	handlers := handler.NewHandlers(client,
		handler.WithNotifyURL(notifyURL),
		handler.WithReturnURL(returnURL),
	)

	// 注册路由 - 标准库 http.Handler，兼容所有框架
	// 包装表单支付处理器以存储订单
	http.HandleFunc("/pay/form", func(w http.ResponseWriter, r *http.Request) {
		// 在调用 handler 前存储订单信息
		payType := r.URL.Query().Get("type")
		name := r.URL.Query().Get("name")
		moneyStr := r.URL.Query().Get("money")

		money := 0.01
		if m, err := fmt.Sscanf(moneyStr, "%f", &money); err != nil || m == 0 {
			money = 0.01
		}

		outTradeNo := fmt.Sprintf("QUICK%d", time.Now().UnixNano())

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

		// 调用原始 handler
		handlers.FormPayment().ServeHTTP(w, r)
	})
	http.Handle("/pay/qrcode", handlers.QRCodePayment())   // 二维码支付（API）
	http.Handle("/pay/query", handlers.QueryOrder())       // 订单查询
	http.Handle("/return", handlers.Return())              // 支付成功跳转

	// 订单列表 API
	http.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		ordersLock.RLock()
		orderList := make([]*Order, 0, len(orders))
		for _, order := range orders {
			orderList = append(orderList, order)
		}
		ordersLock.RUnlock()

		sort.Slice(orderList, func(i, j int) bool {
			return orderList[i].CreateTime.After(orderList[j].CreateTime)
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    orderList,
		})
	})

	// 支付回调 - 处理业务逻辑
	http.Handle("/notify", handlers.Notify(func(data *epay.NotifyData) error {
		// 在这里处理你的业务逻辑
		log.Printf("支付成功: 订单号=%s, 金额=%s, 状态=%s",
			data.OutTradeNo, data.Money, data.TradeStatus)

		// 更新内存中的订单状态
		ordersLock.Lock()
		if order, exists := orders[data.OutTradeNo]; exists {
			order.Status = 1
			order.PayTime = time.Now()
			order.TradeNo = data.TradeNo
		}
		ordersLock.Unlock()

		// TODO: 更新订单状态、发放商品等
		// 返回 nil 表示处理成功，返回 error 会向 EPay 返回 "fail"
		return nil
	}))

	// 简单的首页
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>EPay 快速示例</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1100px; margin: 50px auto; padding: 20px; }
        .container { display: flex; gap: 30px; }
        .left-panel, .right-panel { flex: 1; }
        h1 { color: #333; }
        a { color: #1890ff; }
        .order-list { margin-top: 20px; }
        .order-item { background: #fff; border: 1px solid #e8e8e8; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .order-header { display: flex; justify-content: space-between; margin-bottom: 10px; }
        .order-no { font-weight: bold; color: #333; font-size: 14px; }
        .order-status { padding: 2px 8px; border-radius: 3px; font-size: 12px; }
        .status-paid { background: #52c41a; color: white; }
        .status-unpaid { background: #faad14; color: white; }
        .order-info { font-size: 13px; color: #666; }
        .order-info p { margin: 5px 0; }
        .order-money { font-size: 18px; color: #f5222d; font-weight: bold; }
        .refresh-btn { background: #52c41a; color: white; padding: 5px 10px; border: none; cursor: pointer; margin-left: 10px; }
        .no-orders { text-align: center; color: #999; padding: 40px; }
        @media (max-width: 900px) { .container { flex-direction: column; } }
    </style>
</head>
<body>
    <h1>EPay 快速集成示例</h1>
    <div class="container">
        <div class="left-panel">
            <h3>表单支付（跳转）</h3>
            <ul>
                <li><a href="/pay/form?type=alipay&name=测试商品&money=0.01">支付宝支付</a></li>
                <li><a href="/pay/form?type=wxpay&name=测试商品&money=0.01">微信支付</a></li>
                <li><a href="/pay/form?type=&name=测试商品&money=0.01">收银台</a></li>
            </ul>
            <h3>API 接口</h3>
            <ul>
                <li>POST /pay/qrcode - 创建支付（返回二维码链接）</li>
                <li>GET /pay/query?out_trade_no=xxx - 查询订单</li>
            </ul>
        </div>
        <div class="right-panel">
            <h2>订单列表 <button class="refresh-btn" onclick="loadOrders()">刷新</button></h2>
            <div id="orderList" class="order-list">
                <div class="no-orders">暂无订单</div>
            </div>
        </div>
    </div>
    <script>
        async function loadOrders() {
            try {
                const resp = await fetch('/api/orders');
                const result = await resp.json();
                const el = document.getElementById('orderList');
                if (result.success && result.data && result.data.length > 0) {
                    el.innerHTML = result.data.map(o => {
                        const sc = o.status === 1 ? 'status-paid' : 'status-unpaid';
                        const st = o.status === 1 ? '已支付' : '未支付';
                        const pt = {alipay:'支付宝',wxpay:'微信',qqpay:'QQ钱包','':'收银台'}[o.pay_type] || o.pay_type;
                        const ct = new Date(o.create_time).toLocaleString('zh-CN');
                        return '<div class="order-item"><div class="order-header"><span class="order-no">'+o.out_trade_no+'</span><span class="order-status '+sc+'">'+st+'</span></div><div class="order-info"><p><strong>商品：</strong>'+o.name+'</p><p><strong>金额：</strong><span class="order-money">¥'+o.money.toFixed(2)+'</span></p><p><strong>方式：</strong>'+pt+'</p><p><strong>时间：</strong>'+ct+'</p></div></div>';
                    }).join('');
                } else {
                    el.innerHTML = '<div class="no-orders">暂无订单</div>';
                }
            } catch (err) { console.error(err); }
        }
        loadOrders();
        setInterval(loadOrders, 5000);
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("EPay 快速示例服务启动: http://localhost:%s", port)
	log.Printf("请设置环境变量: EPAY_PID, EPAY_KEY, EPAY_API_URL, EPAY_NOTIFY_URL, EPAY_RETURN_URL")
	log.Printf("当前配置: NotifyURL=%s, ReturnURL=%s", notifyURL, returnURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
