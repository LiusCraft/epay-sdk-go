// Quick 快速启动示例
// 演示如何用几行代码快速集成 EPay 支付
package main

import (
	"log"
	"net/http"
	"os"

	epay "github.com/liuscraft/epay-sdk-go"
	"github.com/liuscraft/epay-sdk-go/handler"
)

func main() {
	// 方式 1: 一行代码创建客户端（使用默认配置）
	client := epay.NewQuick(
		1001,                       // PID
		"your-merchant-key",        // Key
		"https://pay.example.com",  // API URL
	)

	// 方式 2: 链式 API（更多配置选项）
	// client := epay.New(1001, "your-merchant-key", "https://pay.example.com").
	//     WithTimeout(30).
	//     WithDebug(true).
	//     MustBuild()

	// 创建 HTTP Handlers
	handlers := handler.NewHandlers(client,
		handler.WithNotifyURL("https://yourdomain.com/notify"),
		handler.WithReturnURL("https://yourdomain.com/return"),
	)

	// 注册路由 - 标准库 http.Handler，兼容所有框架
	http.Handle("/pay/form", handlers.FormPayment())       // 表单支付
	http.Handle("/pay/qrcode", handlers.QRCodePayment())   // 二维码支付（API）
	http.Handle("/pay/query", handlers.QueryOrder())       // 订单查询
	http.Handle("/return", handlers.Return())              // 支付成功跳转

	// 支付回调 - 处理业务逻辑
	http.Handle("/notify", handlers.Notify(func(data *epay.NotifyData) error {
		// 在这里处理你的业务逻辑
		log.Printf("支付成功: 订单号=%s, 金额=%s, 状态=%s",
			data.OutTradeNo, data.Money, data.TradeStatus)

		// TODO: 更新订单状态、发放商品等
		// 返回 nil 表示处理成功，返回 error 会向 EPay 返回 "fail"
		return nil
	}))

	// 简单的首页
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>EPay 快速示例</title></head>
<body>
    <h1>EPay 快速集成示例</h1>
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
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
