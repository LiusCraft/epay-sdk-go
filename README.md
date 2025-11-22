# EPay Go SDK

EPay（易支付）Go 语言 SDK，支持支付宝、微信支付等多种支付方式。

## 安装

```bash
go get github.com/liuscraft/epay-sdk-go
```

## 快速开始

```go
package main

import (
    "fmt"
    "log"

    epay "github.com/liuscraft/epay-sdk-go"
)

func main() {
    // 创建客户端
    client, err := epay.NewClient(&epay.Config{
        PID:        1001,                          // 商户ID
        Key:        "your-merchant-key",           // 商户密钥
        APIBaseURL: "https://pay.example.com",     // EPay 服务器地址
    })
    if err != nil {
        log.Fatal(err)
    }

    // 创建支付订单
    resp, err := client.CreatePayment(&epay.PaymentRequest{
        Type:       "alipay",
        OutTradeNo: "ORDER20231123001",
        NotifyURL:  "https://yourdomain.com/notify",
        Name:       "商品名称",
        Money:      9.99,
        ClientIP:   "127.0.0.1",
        Device:     "pc",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("支付链接: %s\n", resp.PayURL)
}
```

## 功能特性

- **Form 表单支付** - 页面跳转到支付收银台
- **API 接口支付** - 获取二维码/支付链接
- **支付回调验证** - 验证异步通知签名
- **订单查询** - 查询订单支付状态
- **退款申请** - 提交退款请求

## 支付方式

| 支付方式 | Type 参数 |
|----------|-----------|
| 支付宝 | `alipay` |
| 微信支付 | `wxpay` |
| QQ 钱包 | `qqpay` |

## 示例

查看 [examples](./examples) 目录获取完整示例：

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

详细设计文档请查看 [docs/SDK_DESIGN.md](./docs/SDK_DESIGN.md)

## License

MIT License
