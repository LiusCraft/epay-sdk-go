# EPay Go SDK

EPay（易支付）Go 语言 SDK，支持支付宝、微信支付等多种支付方式。

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
  - [方式 1: 一行代码创建客户端（推荐）](#方式-1-一行代码创建客户端推荐)
  - [方式 2: 链式 API（更多配置）](#方式-2-链式-api更多配置)
  - [方式 3: 传统方式](#方式-3-传统方式)
- [功能特性](#功能特性)
- [支付方式](#支付方式)
- [框架集成](#框架集成)
  - [标准库 net/http](#标准库-nethttp)
  - [Gin 框架](#gin-框架)
  - [Echo 框架](#echo-框架)
  - [Chi 路由](#chi-路由)
- [示例](#示例)
  - [Form 表单支付](#form-表单支付)
  - [API 接口支付](#api-接口支付)
  - [支付回调处理](#支付回调处理)
- [配置说明](#配置说明)
- [错误处理](#错误处理)
- [安全建议](#安全建议)
- [文档](#文档)
- [License](#license)

---

## 安装

```bash
go get github.com/liuscraft/epay-sdk-go
```

## 快速开始

### 方式 1: 一行代码创建客户端（推荐）

```go
package main

import (
    "log"
    "net/http"

    epay "github.com/liuscraft/epay-sdk-go"
    "github.com/liuscraft/epay-sdk-go/handler"
)

func main() {
    // 一行代码创建客户端
    client := epay.NewQuick(1001, "your-key", "https://pay.example.com")

    // 创建 HTTP Handlers（兼容所有框架）
    handlers := handler.NewHandlers(client,
        handler.WithNotifyURL("https://yourdomain.com/notify"),
        handler.WithReturnURL("https://yourdomain.com/return"),
    )

    // 注册路由
    http.Handle("/pay", handlers.FormPayment())
    http.Handle("/notify", handlers.Notify(func(data *epay.NotifyData) error {
        log.Printf("支付成功: %s", data.OutTradeNo)
        return nil
    }))

    http.ListenAndServe(":8080", nil)
}
```

### 方式 2: 链式 API（更多配置）

```go
client := epay.New(1001, "your-key", "https://pay.example.com").
    WithTimeout(30).
    WithDebug(true).
    Build()
```

### 方式 3: 传统方式

```go
client, err := epay.NewClient(&epay.Config{
    PID:        1001,
    Key:        "your-merchant-key",
    APIBaseURL: "https://pay.example.com",
    Timeout:    30,
    Debug:      false,
})
```

## 功能特性

- ✨ **一行代码集成** - `epay.NewQuick()` 快速创建客户端
- 🔗 **链式 API** - 优雅的链式调用方式
- 🎯 **标准 http.Handler** - 兼容所有 Go Web 框架（Gin、Echo、Chi、Fiber 等）
- 📝 **Form 表单支付** - 页面跳转到支付收银台
- 💳 **API 接口支付** - 获取二维码/支付链接
- ✅ **支付回调验证** - 自动验证异步通知签名
- 🔍 **订单查询** - 查询订单支付状态
- 💰 **退款申请** - 提交退款请求
- 🛠️ **开箱即用** - 内置 Handler，无需重复编写路由逻辑

## 支付方式

| 支付方式 | Type 参数 |
|----------|-----------|
| 支付宝 | `alipay` |
| 微信支付 | `wxpay` |
| QQ 钱包 | `qqpay` |

## 框架集成

EPay SDK 提供标准的 `http.Handler`，可以无缝集成到任何 Go Web 框架：

### 标准库 net/http

```go
http.Handle("/pay", handlers.FormPayment())
http.Handle("/notify", handlers.Notify(callback))
```

### Gin 框架

```go
r.GET("/pay", gin.WrapH(handlers.FormPayment()))
r.POST("/notify", gin.WrapH(handlers.Notify(callback)))
```

### Echo 框架

```go
e.GET("/pay", echo.WrapHandler(handlers.FormPayment()))
e.POST("/notify", echo.WrapHandler(handlers.Notify(callback)))
```

### Chi 路由

```go
r.Handle("/pay", handlers.FormPayment())
r.Handle("/notify", handlers.Notify(callback))
```

查看 [Handler 使用指南 - 框架集成](./docs/HANDLER_GUIDE.md#框架集成) 获取更多示例。

## 示例

查看 [examples](./examples) 目录获取完整示例：

- [快速开始示例](./examples/quick/main.go) - 一行代码集成（推荐）
- [Form 表单支付示例](./examples/form/main.go) - 页面跳转支付
- [API 接口支付示例](./examples/api/main.go) - 获取二维码/支付链接

### Form 表单支付

适用于网页端直接跳转到支付页面：

```go
// 生成跳转 URL
payURL, err := client.BuildFormPaymentURL(&epay.FormPaymentRequest{
    OutTradeNo: "ORDER001",
    NotifyURL:  "https://yourdomain.com/notify",
    ReturnURL:  "https://yourdomain.com/success",
    Name:       "VIP会员",
    Money:      99.00,
})

// 或生成 HTML 表单
htmlForm, err := client.BuildFormPayment(&epay.FormPaymentRequest{...})
```

### API 接口支付

适用于获取二维码展示给用户：

```go
resp, err := client.CreatePayment(&epay.PaymentRequest{
    Type:       "wxpay",
    OutTradeNo: "ORDER001",
    NotifyURL:  "https://yourdomain.com/notify",
    Name:       "商品名称",
    Money:      9.99,
    ClientIP:   "127.0.0.1",
    Device:     "pc",
})

// resp.QRCode - 二维码链接
// resp.PayURL - 支付跳转链接
```

### 支付回调处理

```go
func notifyHandler(w http.ResponseWriter, r *http.Request) {
    params := epay.ParseNotifyParams(r)

    notifyData, err := client.VerifyNotify(params)
    if err != nil {
        w.Write([]byte("fail"))
        return
    }

    if notifyData.TradeStatus == "TRADE_SUCCESS" {
        // 处理业务逻辑
    }

    w.Write([]byte("success"))
}
```

## 配置说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| PID | int | 是 | 商户ID |
| Key | string | 是 | 商户密钥 |
| APIBaseURL | string | 是 | EPay 服务器地址 |
| Timeout | int | 否 | 请求超时（秒），默认 30 |
| Debug | bool | 否 | 调试模式，默认 false |

## 错误处理

```go
resp, err := client.CreatePayment(req)
if err != nil {
    if epayErr, ok := err.(*epay.EPayError); ok {
        switch epayErr.Code {
        case epay.ErrCodeSignFailed:
            // 签名错误
        case epay.ErrCodeAPIError:
            // API 错误
        case epay.ErrCodeNetworkError:
            // 网络错误
        }
    }
}
```

## 安全建议

1. **商户密钥** - 使用环境变量存储，不要硬编码
2. **回调验证** - 必须验证签名，防止伪造请求
3. **幂等处理** - 回调可能重复，需要幂等性处理
4. **HTTPS** - 生产环境必须使用 HTTPS

## 文档

- [Handler 使用指南](./docs/HANDLER_GUIDE.md) - 详细说明每个 Handler 的作用和使用方法（包含框架集成示例）
- [SDK 设计文档](./docs/SDK_DESIGN.md) - SDK 架构设计和实现细节

## License

MIT License
