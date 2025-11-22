# Handler 使用指南

本文档详细说明 `handler` 包提供的所有 HTTP Handler 及其使用方法。

## 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [Handlers 配置](#handlers-配置)
  - [NewHandlers](#newhandlers)
  - [配置选项](#配置选项)
    - [WithNotifyURL](#withnotifyurl)
    - [WithReturnURL](#withreturnurl)
    - [WithLogger](#withlogger)
- [Handler 详解](#handler-详解)
  - [1. FormPayment - 表单支付](#1-formpayment)
  - [2. QRCodePayment - 二维码支付](#2-qrcodepayment)
  - [3. Notify - 支付回调](#3-notify)
  - [4. Return - 同步跳转](#4-return)
  - [5. QueryOrder - 订单查询](#5-queryorder)
- [完整示例](#完整示例)
  - [最小化示例](#最小化示例10-行代码)
  - [完整示例](#完整示例包含所有-handler)
- [框架集成](#框架集成)
  - [Gin 框架](#gin-框架)
  - [Echo 框架](#echo-框架)
- [最佳实践](#最佳实践)
  - [1. 回调处理](#1-回调处理notify)
  - [2. 订单号生成](#2-订单号生成)
  - [3. 生产环境配置](#3-生产环境配置)
- [常见问题](#常见问题)
- [更多资源](#更多资源)

---

## 概述

`handler` 包提供了一套标准的 `http.Handler` 实现，开箱即用，兼容所有 Go Web 框架。开发者无需自己编写路由处理逻辑，直接注册到框架即可。

## 快速开始

```go
import (
    epay "github.com/liuscraft/epay-sdk-go"
    "github.com/liuscraft/epay-sdk-go/handler"
)

// 1. 创建客户端
client := epay.NewQuick(1001, "your-key", "https://pay.example.com")

// 2. 创建 Handlers
handlers := handler.NewHandlers(client,
    handler.WithNotifyURL("https://yourdomain.com/notify"),
    handler.WithReturnURL("https://yourdomain.com/return"),
)

// 3. 注册路由
http.Handle("/pay/form", handlers.FormPayment())
http.Handle("/pay/qrcode", handlers.QRCodePayment())
http.Handle("/pay/query", handlers.QueryOrder())
http.Handle("/notify", handlers.Notify(callback))
http.Handle("/return", handlers.Return())
```

## Handlers 配置

### NewHandlers

创建 Handler 集合。

**函数签名：**
```go
func NewHandlers(client *epay.Client, opts ...Option) *Handlers
```

**参数：**
- `client` - EPay 客户端实例
- `opts` - 配置选项（可选）

### 配置选项

#### WithNotifyURL

设置异步回调地址（EPay 服务器会向此地址发送支付结果）。

```go
handler.WithNotifyURL("https://yourdomain.com/api/payment/notify")
```

**说明：**
- 必须是可公网访问的 HTTPS 地址
- EPay 服务器会 POST 支付结果到此地址
- 如果不设置，创建支付时需要手动指定

#### WithReturnURL

设置同步跳转地址（用户支付完成后跳转的页面）。

```go
handler.WithReturnURL("https://yourdomain.com/payment/success")
```

**说明：**
- 用户支付完成后会跳转到此页面
- 可以展示支付结果给用户
- 不保证一定会跳转（用户可能关闭页面）

#### WithLogger

设置自定义日志器。

```go
handler.WithLogger(customLogger)
```

**说明：**
- 默认使用 `log.Default()`
- 可以传入自定义 Logger 实现

**示例：**
```go
handlers := handler.NewHandlers(client,
    handler.WithNotifyURL("https://yourdomain.com/notify"),
    handler.WithReturnURL("https://yourdomain.com/return"),
    handler.WithLogger(myLogger),
)
```

---

## Handler 详解

### 1. FormPayment

**用途：** 表单支付（页面跳转）

**适用场景：**
- PC 网站支付
- H5 移动网页支付
- 需要用户在收银台选择支付方式

**HTTP 方法：** GET

**URL 参数：**
| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| type | string | 否 | 支付方式：alipay/wxpay/qqpay，为空则显示收银台 | `alipay` |
| name | string | 否 | 商品名称，默认"商品" | `VIP会员` |
| money | float | 是 | 金额（元），必须大于 0 | `0.01` |

**返回：**
- Content-Type: `text/html`
- 自动提交的 HTML 表单，浏览器会自动跳转到支付页面

**示例：**
```go
// 注册路由
http.Handle("/pay/form", handlers.FormPayment())

// 用户访问
// https://yourdomain.com/pay/form?type=alipay&name=VIP会员&money=99.00
// -> 自动跳转到支付宝支付页面

// https://yourdomain.com/pay/form?name=测试商品&money=0.01
// -> 显示收银台，用户选择支付方式
```

**自动生成订单号：**
- 格式：`ORDER{时间戳纳秒}`
- 示例：`ORDER1732345678901234567`

---

### 2. QRCodePayment

**用途：** API 接口支付（返回二维码/支付链接）

**适用场景：**
- 前后端分离项目
- 需要展示二维码给用户扫码支付
- 需要获取支付链接用于其他用途

**HTTP 方法：** POST

**请求头：**
```
Content-Type: application/json
```

**请求体（JSON）：**
```json
{
    "pay_type": "alipay",    // 支付方式：alipay/wxpay/qqpay
    "name": "商品名称",       // 商品名称
    "money": 0.01            // 金额（元）
}
```

**响应（JSON）：**

成功：
```json
{
    "success": true,
    "data": {
        "out_trade_no": "API1732345678901234567",  // 商户订单号
        "trade_no": "20231123001",                  // EPay 订单号
        "pay_url": "https://pay.example.com/...",   // 支付链接
        "qr_code": "https://qr.example.com/...",    // 二维码链接
        "url_scheme": "alipays://..."               // URL Scheme（移动端）
    }
}
```

失败：
```json
{
    "success": false,
    "message": "Invalid money"  // 错误信息
}
```

**示例：**
```go
// 注册路由
http.Handle("/pay/qrcode", handlers.QRCodePayment())

// 前端调用
fetch('/pay/qrcode', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        pay_type: 'alipay',
        name: '测试商品',
        money: 0.01
    })
}).then(res => res.json())
  .then(data => {
      if (data.success) {
          // 展示二维码：data.data.qr_code
          // 或跳转支付：window.location.href = data.data.pay_url
      }
  })
```

**自动生成订单号：**
- 格式：`API{时间戳纳秒}`
- 示例：`API1732345678901234567`

---

### 3. Notify

**用途：** 处理支付异步回调

**适用场景：**
- 接收 EPay 服务器的支付结果通知
- 更新订单状态
- 发放商品/服务

**HTTP 方法：** POST（EPay 服务器调用）

**参数：** 由 EPay 服务器自动发送（表单格式）

**返回：**
- `success` - 处理成功
- `fail` - 处理失败（EPay 会重试）

**函数签名：**
```go
func (h *Handlers) Notify(callback NotifyCallback) http.Handler
```

**回调函数类型：**
```go
type NotifyCallback func(notifyData *epay.NotifyData) error
```

**NotifyData 结构：**
```go
type NotifyData struct {
    PID         string  // 商户 ID
    TradeNo     string  // EPay 订单号
    OutTradeNo  string  // 商户订单号
    Type        string  // 支付方式
    Name        string  // 商品名称
    Money       string  // 支付金额
    TradeStatus string  // 交易状态：TRADE_SUCCESS
    Param       string  // 自定义参数
}
```

**示例：**
```go
// 定义回调处理函数
func handleNotify(data *epay.NotifyData) error {
    log.Printf("收到支付通知: 订单=%s, 金额=%s", data.OutTradeNo, data.Money)

    // 1. 检查订单状态
    if data.TradeStatus != "TRADE_SUCCESS" {
        return fmt.Errorf("订单未支付")
    }

    // 2. 检查订单是否已处理（防止重复处理）
    if orderAlreadyProcessed(data.OutTradeNo) {
        return nil  // 已处理，直接返回成功
    }

    // 3. 验证金额是否正确
    expectedAmount := getOrderAmount(data.OutTradeNo)
    if data.Money != fmt.Sprintf("%.2f", expectedAmount) {
        return fmt.Errorf("金额不匹配")
    }

    // 4. 更新订单状态
    if err := updateOrderStatus(data.OutTradeNo, "paid"); err != nil {
        return err  // 返回 error，EPay 会重试
    }

    // 5. 发放商品/服务
    if err := deliverGoods(data.OutTradeNo); err != nil {
        return err
    }

    // 6. 记录日志
    logPaymentSuccess(data)

    return nil  // 返回 nil 表示处理成功
}

// 注册路由
http.Handle("/notify", handlers.Notify(handleNotify))
```

**重要说明：**

1. **签名验证：** Handler 自动验证签名，无需手动验证
2. **幂等性：** 回调可能重复，必须做幂等处理
3. **错误处理：**
   - 返回 `nil` - 向 EPay 返回 "success"，EPay 不再重试
   - 返回 `error` - 向 EPay 返回 "fail"，EPay 会重试
4. **重试机制：** EPay 会重试通知，间隔：5s、15s、30s、1min、2min、5min...
5. **超时时间：** 回调处理应在 30 秒内完成

---

### 4. Return

**用途：** 处理支付同步跳转

**适用场景：**
- 用户支付完成后的跳转页面
- 展示支付结果给用户

**HTTP 方法：** GET

**URL 参数：** 由 EPay 自动附加

**返回：**
- Content-Type: `text/html`
- 简单的支付成功页面

**示例：**
```go
// 注册路由
http.Handle("/return", handlers.Return())

// 用户支付完成后会跳转到此页面
// https://yourdomain.com/return?out_trade_no=xxx&money=0.01&...
```

**自定义返回页面：**

如果需要自定义页面，可以自己实现 Handler：

```go
http.HandleFunc("/return", func(w http.ResponseWriter, r *http.Request) {
    params := epay.ParseNotifyParams(r)

    // 渲染自定义模板
    tmpl.Execute(w, map[string]interface{}{
        "OrderNo": params["out_trade_no"],
        "Amount":  params["money"],
    })
})
```

**注意：**
- 同步跳转不保证一定会调用（用户可能关闭页面）
- 不要依赖同步跳转更新订单状态
- 订单状态更新应在异步回调（Notify）中处理

---

### 5. QueryOrder

**用途：** 查询订单状态

**适用场景：**
- 主动查询订单支付状态
- 前端轮询支付结果
- 后台管理查询订单

**HTTP 方法：** GET

**URL 参数（二选一）：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| out_trade_no | string | 否 | 商户订单号 |
| trade_no | string | 否 | EPay 订单号 |

**响应（JSON）：**

成功：
```json
{
    "success": true,
    "data": {
        "out_trade_no": "ORDER123",    // 商户订单号
        "trade_no": "20231123001",     // EPay 订单号
        "type": "alipay",              // 支付方式
        "name": "测试商品",             // 商品名称
        "money": "0.01",               // 金额
        "status": 1,                   // 状态：1=已支付, 0=未支付
        "add_time": "2023-11-23 10:00:00",   // 创建时间
        "end_time": "2023-11-23 10:05:00"    // 支付时间
    }
}
```

失败：
```json
{
    "success": false,
    "message": "out_trade_no or trade_no is required"
}
```

**示例：**
```go
// 注册路由
http.Handle("/pay/query", handlers.QueryOrder())

// 前端轮询查询
async function checkPaymentStatus(outTradeNo) {
    const resp = await fetch(`/pay/query?out_trade_no=${outTradeNo}`)
    const data = await resp.json()

    if (data.success && data.data.status === 1) {
        // 支付成功
        showSuccessMessage()
    } else {
        // 继续轮询
        setTimeout(() => checkPaymentStatus(outTradeNo), 3000)
    }
}
```

---

## 完整示例

### 最小化示例（10 行代码）

```go
package main

import (
    "log"
    "net/http"
    epay "github.com/liuscraft/epay-sdk-go"
    "github.com/liuscraft/epay-sdk-go/handler"
)

func main() {
    client := epay.NewQuick(1001, "key", "https://pay.example.com")
    handlers := handler.NewHandlers(client,
        handler.WithNotifyURL("https://yourdomain.com/notify"),
    )

    http.Handle("/pay", handlers.FormPayment())
    http.Handle("/notify", handlers.Notify(func(data *epay.NotifyData) error {
        log.Printf("支付成功: %s", data.OutTradeNo)
        return nil
    }))

    http.ListenAndServe(":8080", nil)
}
```

### 完整示例（包含所有 Handler）

参见 `examples/quick/main.go`

---

## 框架集成

### Gin 框架

```go
r := gin.Default()
r.GET("/pay", gin.WrapH(handlers.FormPayment()))
r.POST("/pay/qrcode", gin.WrapH(handlers.QRCodePayment()))
r.POST("/notify", gin.WrapH(handlers.Notify(callback)))
```

### Echo 框架

```go
e := echo.New()
e.GET("/pay", echo.WrapHandler(handlers.FormPayment()))
e.POST("/pay/qrcode", echo.WrapHandler(handlers.QRCodePayment()))
e.POST("/notify", echo.WrapHandler(handlers.Notify(callback)))
```

---

## 最佳实践

### 1. 回调处理（Notify）

✅ **正确做法：**
```go
func handleNotify(data *epay.NotifyData) error {
    // 1. 检查是否已处理（幂等性）
    if isProcessed(data.OutTradeNo) {
        return nil
    }

    // 2. 验证金额
    if !verifyAmount(data.OutTradeNo, data.Money) {
        return fmt.Errorf("金额不匹配")
    }

    // 3. 使用事务更新订单
    tx := db.Begin()
    if err := tx.UpdateOrder(data.OutTradeNo, "paid"); err != nil {
        tx.Rollback()
        return err
    }
    tx.Commit()

    // 4. 异步发放商品
    go deliverGoods(data.OutTradeNo)

    return nil
}
```

❌ **错误做法：**
```go
func handleNotify(data *epay.NotifyData) error {
    // 不检查幂等性 - 可能重复处理！
    updateOrder(data.OutTradeNo, "paid")

    // 不验证金额 - 可能被篡改！

    // 总是返回 nil - 即使出错也不重试！
    return nil
}
```

### 2. 订单号生成

Handler 会自动生成订单号，但建议使用自己的订单号：

```go
// 自定义实现，不使用 Handler
http.HandleFunc("/pay/custom", func(w http.ResponseWriter, r *http.Request) {
    outTradeNo := generateOrderNo()  // 自己生成

    htmlForm, err := client.BuildFormPayment(&epay.FormPaymentRequest{
        OutTradeNo: outTradeNo,
        // ...
    })

    w.Write([]byte(htmlForm))
})
```

### 3. 生产环境配置

```go
handlers := handler.NewHandlers(client,
    // 使用环境变量
    handler.WithNotifyURL(os.Getenv("EPAY_NOTIFY_URL")),
    handler.WithReturnURL(os.Getenv("EPAY_RETURN_URL")),
    // 使用自定义 Logger
    handler.WithLogger(myLogger),
)
```

---

## 常见问题

### Q: Handler 生成的订单号格式是什么？
A:
- FormPayment: `ORDER{时间戳纳秒}`
- QRCodePayment: `API{时间戳纳秒}`

### Q: 可以自定义订单号吗？
A: 可以。不使用 Handler，直接调用 Client 方法，自己实现 Handler。

### Q: Notify 回调会调用多次吗？
A: 是的。EPay 会重试直到收到 "success"，所以必须做幂等处理。

### Q: 签名验证失败怎么办？
A: Handler 会自动返回 "fail"。检查商户密钥是否正确。

### Q: 如何自定义支付成功页面？
A: 不使用 `handlers.Return()`，自己实现页面。

### Q: 支持中间件吗？
A: 支持。Handler 返回标准 `http.Handler`，可以用任何中间件包装。

```go
// 使用中间件
http.Handle("/pay", LoggingMiddleware(handlers.FormPayment()))
```

---

## 更多资源

- [快速开始示例](../examples/quick/main.go)
- [SDK 设计文档](./SDK_DESIGN.md)
- [API 参考文档](./API_REFERENCE.md)（待补充）
