# Epay Go SDK 设计文档

> **文档版本**: v1.1
> **创建日期**: 2025-11-23
> **最后更新**: 2025-11-23
> **仓库地址**: github.com/liuscraft/epay-sdk-go

---

## 目录

- [1. 概述](#1-概述)
- [2. 快速开始](#2-快速开始)
- [3. 架构设计](#3-架构设计)
- [4. 核心功能模块](#4-核心功能模块)
- [5. 数据结构定义](#5-数据结构定义)
- [6. 接口设计](#6-接口设计)
- [7. 签名算法实现](#7-签名算法实现)
- [8. 错误处理](#8-错误处理)
- [9. 使用示例](#9-使用示例)
- [10. 安全性考虑](#10-安全性考虑)
- [11. 测试计划](#11-测试计划)

---

## 1. 概述

### 1.1 背景

EPay（易支付）是一个第三方支付平台，支持支付宝、微信支付等多种支付方式。本 SDK 旨在为 Go 开发者提供便捷的支付集成能力。

### 1.2 设计目标

- **易用性**: 简洁的 API 设计，快速集成
- **安全性**: 完善的签名验证机制
- **可维护性**: 清晰的代码结构，易于扩展
- **健壮性**: 完善的错误处理和重试机制
- **可测试性**: 支持单元测试和集成测试

### 1.3 技术栈

- **语言**: Go 1.21+
- **HTTP 客户端**: 标准库 `net/http`
- **JSON 处理**: 标准库 `encoding/json`
- **签名算法**: MD5（`crypto/md5`）
- **日志**: 可选接入项目日志系统

---

## 2. 快速开始

### 2.1 安装

```bash
go get github.com/liuscraft/epay-sdk-go
```

### 2.2 基础使用

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
        PID:        1001,
        Key:        "your-merchant-key",
        APIBaseURL: "https://your-epay-server.com",
        Timeout:    30,
        Debug:      true,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
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
        log.Fatalf("Payment failed: %v", err)
    }

    fmt.Printf("支付链接: %s\n", resp.PayURL)
}
```

---

## 3. 架构设计

### 3.1 整体架构

```
┌─────────────────────────────────────────────────┐
│              Your Application                   │
│         (调用 EPay SDK 发起支付)                │
└─────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────┐
│          github.com/liuscraft/epay-sdk-go       │
├─────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │  Client  │  │  Signer  │  │ Verifier │      │
│  │  (核心)  │  │ (签名器) │  │ (验证器) │      │
│  └──────────┘  └──────────┘  └──────────┘      │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │  Models  │  │  Errors  │  │  Utils   │      │
│  │ (数据结构)│  │ (错误类型)│  │ (工具类) │      │
│  └──────────┘  └──────────┘  └──────────┘      │
└─────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────┐
│              EPay API Server                    │
└─────────────────────────────────────────────────┘
```

### 3.2 目录结构

```
epay-sdk-go/
├── client.go          # SDK 客户端核心实现
├── config.go          # 配置结构
├── signer.go          # MD5 签名器
├── verifier.go        # 签名验证器
├── models.go          # 数据模型定义
├── errors.go          # 错误定义
├── utils.go           # 工具函数
├── payment.go         # 支付相关接口
├── order.go           # 订单查询接口
├── refund.go          # 退款接口
├── client_test.go     # 单元测试
├── go.mod             # Go 模块定义
├── README.md          # 项目说明
├── docs/
│   └── SDK_DESIGN.md  # 设计文档
└── examples/
    ├── form/
    │   └── main.go    # 表单跳转支付示例
    └── api/
        └── main.go    # API 接口支付示例
```

### 3.3 核心组件职责

| 组件 | 职责 |
|------|------|
| **Client** | SDK 主入口，管理配置、HTTP 请求、API 调用 |
| **Signer** | 实现 MD5 签名算法 |
| **Verifier** | 验证回调通知签名 |
| **Models** | 定义请求/响应数据结构 |
| **Errors** | 统一错误类型和错误码 |
| **Utils** | 通用工具函数（参数排序、URL 编码等） |

---

## 4. 核心功能模块

### 4.1 功能列表

| 功能模块 | 优先级 | 说明 |
|----------|--------|------|
| **支付发起** | P0 | 页面跳转支付（Form）、API接口支付 |
| **回调处理** | P0 | 异步通知、同步跳转验证 |
| **订单查询** | P1 | 单个订单查询、批量订单查询 |
| **订单退款** | P1 | 提交退款申请 |
| **签名验证** | P0 | MD5签名、签名验证 |

### 4.2 支付方式对比

| 方式 | 适用场景 | 返回内容 |
|------|----------|----------|
| **Form 表单支付** | 网页端直接跳转 | HTML 表单或跳转 URL |
| **API 接口支付** | 服务端调用 | 二维码/支付链接 |

### 4.3 支付流程

```
用户下单
   ↓
后端调用 SDK 创建支付订单
   ↓
SDK 生成签名并请求 EPay API
   ↓
EPay 返回支付链接/二维码
   ↓
前端展示支付页面/二维码
   ↓
用户完成支付
   ↓
EPay 发送异步通知到 notify_url
   ↓
SDK 验证签名并返回业务数据
   ↓
业务层处理订单完成逻辑
   ↓
返回 "success" 给 EPay
```

---

## 5. 数据结构定义

### 5.1 配置结构

```go
// Config EPay SDK 配置
type Config struct {
    PID        int    // 商户ID
    Key        string // 商户密钥
    APIBaseURL string // API 基础URL（如: https://pay.example.com）
    Timeout    int    // 请求超时时间（秒，默认: 30）
    Debug      bool   // 是否开启调试模式
}
```

### 5.2 支付请求结构

```go
// PaymentRequest API 接口支付请求
type PaymentRequest struct {
    Type       string  // 支付方式: alipay, wxpay, qqpay 等
    OutTradeNo string  // 商户订单号（唯一）
    NotifyURL  string  // 异步通知地址
    ReturnURL  string  // 同步跳转地址（可选）
    Name       string  // 商品名称
    Money      float64 // 商品金额（元）
    ClientIP   string  // 用户IP地址
    Device     string  // 设备类型: pc, mobile, wechat, alipay
    Param      string  // 业务扩展参数（可选）
}

// FormPaymentRequest 页面跳转支付请求
type FormPaymentRequest struct {
    Type       string  // 支付方式（可选，不传则跳转收银台）
    OutTradeNo string  // 商户订单号
    NotifyURL  string  // 异步通知地址
    ReturnURL  string  // 同步跳转地址
    Name       string  // 商品名称
    Money      float64 // 商品金额
    Param      string  // 业务扩展参数（可选）
}
```

### 5.3 支付响应结构

```go
// PaymentResponse API 接口支付响应
type PaymentResponse struct {
    Code      int    `json:"code"`       // 1=成功，其他=失败
    Msg       string `json:"msg"`        // 错误信息
    TradeNo   string `json:"trade_no"`   // 支付订单号
    PayURL    string `json:"payurl"`     // 支付跳转URL
    QRCode    string `json:"qrcode"`     // 二维码链接
    URLScheme string `json:"urlscheme"`  // 小程序跳转URL
}
```

### 5.4 回调通知结构

```go
// NotifyData 支付回调通知数据
type NotifyData struct {
    PID         int    // 商户ID
    TradeNo     string // EPay订单号
    OutTradeNo  string // 商户订单号
    Type        string // 支付方式
    Name        string // 商品名称
    Money       string // 商品金额
    TradeStatus string // 支付状态（TRADE_SUCCESS）
    Param       string // 业务扩展参数
    Sign        string // 签名字符串
    SignType    string // 签名类型
}
```

### 5.5 订单查询结构

```go
// OrderQueryRequest 订单查询请求
type OrderQueryRequest struct {
    TradeNo    string // EPay订单号（二选一）
    OutTradeNo string // 商户订单号（二选一）
}

// OrderDetail 订单详情
type OrderDetail struct {
    Code       int    `json:"code"`
    Msg        string `json:"msg"`
    TradeNo    string `json:"trade_no"`
    OutTradeNo string `json:"out_trade_no"`
    APITradeNo string `json:"api_trade_no"`
    Type       string `json:"type"`
    PID        int    `json:"pid"`
    AddTime    string `json:"addtime"`
    EndTime    string `json:"endtime"`
    Name       string `json:"name"`
    Money      string `json:"money"`
    Status     int    `json:"status"` // 1=已支付, 0=未支付
    Param      string `json:"param"`
    Buyer      string `json:"buyer"`
}
```

### 5.6 退款请求结构

```go
// RefundRequest 退款请求
type RefundRequest struct {
    TradeNo    string  // EPay订单号（二选一）
    OutTradeNo string  // 商户订单号（二选一）
    Money      float64 // 退款金额
}

// RefundResponse 退款响应
type RefundResponse struct {
    Code int    `json:"code"` // 1=成功
    Msg  string `json:"msg"`
}
```

---

## 6. 接口设计

### 6.1 客户端初始化

```go
// NewClient 创建 EPay 客户端
func NewClient(config *Config) (*Client, error)

// 示例
client, err := epay.NewClient(&epay.Config{
    PID:        1001,
    Key:        "your-merchant-key",
    APIBaseURL: "https://pay.example.com",
    Timeout:    30,
    Debug:      true,
})
```

### 6.2 支付接口

```go
// CreatePayment 创建 API 接口支付
func (c *Client) CreatePayment(req *PaymentRequest) (*PaymentResponse, error)

// BuildFormPayment 构建页面跳转支付 HTML 表单
func (c *Client) BuildFormPayment(req *FormPaymentRequest) (string, error)

// BuildFormPaymentURL 构建页面跳转支付 URL
func (c *Client) BuildFormPaymentURL(req *FormPaymentRequest) (string, error)
```

### 6.3 回调处理接口

```go
// VerifyNotify 验证支付回调通知
func (c *Client) VerifyNotify(params map[string]string) (*NotifyData, error)

// ParseNotifyParams 解析回调参数（从 HTTP Request）
func ParseNotifyParams(r *http.Request) map[string]string
```

### 6.4 订单查询接口

```go
// QueryOrder 查询单个订单
func (c *Client) QueryOrder(req *OrderQueryRequest) (*OrderDetail, error)

// QueryOrders 批量查询订单
func (c *Client) QueryOrders(limit, page int) (*OrderListResponse, error)
```

### 6.5 退款接口

```go
// Refund 提交订单退款
func (c *Client) Refund(req *RefundRequest) (*RefundResponse, error)
```

---

## 7. 签名算法实现

### 7.1 签名生成流程

```go
// Sign 生成 MD5 签名
// 1. 过滤参数：去除 sign、sign_type、空值
// 2. 按 ASCII 排序参数名
// 3. 拼接为 a=b&c=d 格式
// 4. 拼接商户密钥 KEY
// 5. MD5 加密并转小写
func (s *Signer) Sign(params map[string]string) string
```

### 7.2 签名验证流程

```go
// Verify 验证签名是否正确
func (v *Verifier) Verify(params map[string]string, receivedSign string) bool
```

### 7.3 参数处理工具

```go
// SortAndBuildQuery 排序并构建查询字符串
func SortAndBuildQuery(params map[string]string) string

// FilterEmptyParams 过滤空值和特殊参数
func FilterEmptyParams(params map[string]string) map[string]string
```

---

## 8. 错误处理

### 8.1 错误类型定义

```go
// 错误码常量
const (
    ErrCodeInvalidConfig   = 1001 // 配置错误
    ErrCodeSignFailed      = 1002 // 签名失败
    ErrCodeVerifyFailed    = 1003 // 验证失败
    ErrCodeAPIError        = 1004 // API 错误
    ErrCodeNetworkError    = 1005 // 网络错误
    ErrCodeInvalidResponse = 1006 // 响应格式错误
)

// EPayError SDK 错误
type EPayError struct {
    Code    int
    Message string
    Err     error
}

func (e *EPayError) Error() string
```

### 8.2 错误处理最佳实践

```go
import epay "github.com/liuscraft/epay-sdk-go"

resp, err := client.CreatePayment(req)
if err != nil {
    if epayErr, ok := err.(*epay.EPayError); ok {
        switch epayErr.Code {
        case epay.ErrCodeSignFailed:
            // 处理签名错误
        case epay.ErrCodeAPIError:
            // 处理 API 错误
        default:
            // 其他错误
        }
    }
}
```

---

## 9. 使用示例

### 9.1 Form 表单支付（页面跳转）

适用于网页端用户直接跳转到支付页面的场景。

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    epay "github.com/liuscraft/epay-sdk-go"
)

var client *epay.Client

func init() {
    var err error
    client, err = epay.NewClient(&epay.Config{
        PID:        1001,
        Key:        "your-merchant-key",
        APIBaseURL: "https://pay.example.com",
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
}

// 创建支付页面
func createPaymentHandler(w http.ResponseWriter, r *http.Request) {
    // 方式1: 生成 HTML 表单（自动提交）
    htmlForm, err := client.BuildFormPayment(&epay.FormPaymentRequest{
        Type:       "alipay",                                   // 可选，不传则显示收银台
        OutTradeNo: "ORDER" + fmt.Sprint(time.Now().UnixNano()),
        NotifyURL:  "https://yourdomain.com/api/payment/notify",
        ReturnURL:  "https://yourdomain.com/payment/success",
        Name:       "VIP会员",
        Money:      99.00,
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // 返回 HTML 表单，浏览器会自动提交跳转
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(htmlForm))
}

// 或者方式2: 生成跳转 URL
func getPaymentURLHandler(w http.ResponseWriter, r *http.Request) {
    payURL, err := client.BuildFormPaymentURL(&epay.FormPaymentRequest{
        OutTradeNo: "ORDER" + fmt.Sprint(time.Now().UnixNano()),
        NotifyURL:  "https://yourdomain.com/api/payment/notify",
        ReturnURL:  "https://yourdomain.com/payment/success",
        Name:       "VIP会员",
        Money:      99.00,
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // 重定向到支付页面
    http.Redirect(w, r, payURL, http.StatusFound)
}

// 支付回调处理
func notifyHandler(w http.ResponseWriter, r *http.Request) {
    params := epay.ParseNotifyParams(r)

    notifyData, err := client.VerifyNotify(params)
    if err != nil {
        log.Printf("验证失败: %v", err)
        w.Write([]byte("fail"))
        return
    }

    if notifyData.TradeStatus == "TRADE_SUCCESS" {
        // TODO: 处理业务逻辑（更新订单状态等）
        log.Printf("订单支付成功: %s, 金额: %s", notifyData.OutTradeNo, notifyData.Money)
    }

    w.Write([]byte("success"))
}

func main() {
    http.HandleFunc("/pay", createPaymentHandler)
    http.HandleFunc("/pay/url", getPaymentURLHandler)
    http.HandleFunc("/api/payment/notify", notifyHandler)

    log.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

### 9.2 API 接口支付

适用于需要获取二维码或支付链接，由前端展示的场景。

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    epay "github.com/liuscraft/epay-sdk-go"
)

var client *epay.Client

func init() {
    var err error
    client, err = epay.NewClient(&epay.Config{
        PID:        1001,
        Key:        "your-merchant-key",
        APIBaseURL: "https://pay.example.com",
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
}

// API 创建支付订单
func createPaymentAPI(w http.ResponseWriter, r *http.Request) {
    // 解析请求
    var req struct {
        PayType string  `json:"pay_type"`
        Amount  float64 `json:"amount"`
        Name    string  `json:"name"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // 调用 SDK 创建支付
    resp, err := client.CreatePayment(&epay.PaymentRequest{
        Type:       req.PayType,
        OutTradeNo: "ORDER" + fmt.Sprint(time.Now().UnixNano()),
        NotifyURL:  "https://yourdomain.com/api/payment/notify",
        ReturnURL:  "https://yourdomain.com/payment/success",
        Name:       req.Name,
        Money:      req.Amount,
        ClientIP:   r.RemoteAddr,
        Device:     "pc",
        Param:      `{"user_id": "12345"}`, // 可选：业务参数
    })
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    // 返回支付信息给前端
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":  true,
        "trade_no": resp.TradeNo,
        "pay_url":  resp.PayURL,   // 支付跳转链接
        "qr_code":  resp.QRCode,   // 二维码链接（可生成二维码图片）
    })
}

// 查询订单状态
func queryOrderAPI(w http.ResponseWriter, r *http.Request) {
    outTradeNo := r.URL.Query().Get("out_trade_no")

    order, err := client.QueryOrder(&epay.OrderQueryRequest{
        OutTradeNo: outTradeNo,
    })
    if err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":      true,
        "out_trade_no": order.OutTradeNo,
        "trade_no":     order.TradeNo,
        "status":       order.Status, // 1=已支付, 0=未支付
        "money":        order.Money,
    })
}

// 支付回调处理
func notifyHandler(w http.ResponseWriter, r *http.Request) {
    params := epay.ParseNotifyParams(r)

    notifyData, err := client.VerifyNotify(params)
    if err != nil {
        log.Printf("验证失败: %v", err)
        w.Write([]byte("fail"))
        return
    }

    if notifyData.TradeStatus == "TRADE_SUCCESS" {
        // TODO: 处理业务逻辑
        log.Printf("订单支付成功: %s", notifyData.OutTradeNo)
    }

    w.Write([]byte("success"))
}

func main() {
    http.HandleFunc("/api/payment/create", createPaymentAPI)
    http.HandleFunc("/api/payment/query", queryOrderAPI)
    http.HandleFunc("/api/payment/notify", notifyHandler)

    log.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

### 9.3 退款申请

```go
// 提交退款
refund, err := client.Refund(&epay.RefundRequest{
    OutTradeNo: "ORDER20231123001",
    Money:      99.00,
})

if err != nil {
    log.Printf("Refund failed: %v", err)
    return
}

if refund.Code == 1 {
    log.Println("Refund successful")
}
```

---

## 10. 安全性考虑

### 10.1 签名安全

- 商户密钥（Key）必须使用环境变量存储，不得硬编码
- 回调通知必须验证签名，防止伪造
- 签名参数必须按 ASCII 排序，防止篡改

### 10.2 订单安全

- 商户订单号（OutTradeNo）必须唯一，建议使用 UUID 或雪花算法
- 订单金额必须验证，防止金额篡改
- 回调通知需要幂等性处理，防止重复支付

### 10.3 网络安全

- 生产环境必须使用 HTTPS
- 回调 URL（NotifyURL）必须是公网可访问的 HTTPS 地址
- 设置合理的请求超时时间

### 10.4 敏感信息保护

- 日志中不得输出商户密钥
- 用户支付信息需要加密存储
- 定期轮换商户密钥

---

## 11. 测试计划

### 11.1 单元测试

```go
func TestNewClient(t *testing.T)
func TestSign(t *testing.T)
func TestVerify(t *testing.T)
func TestCreatePayment(t *testing.T)
```

### 11.2 集成测试

- 测试完整支付流程（沙箱环境）
- 测试回调通知处理
- 测试订单查询
- 测试退款流程

### 11.3 测试用例

| 测试场景 | 预期结果 |
|----------|----------|
| 正常创建支付订单 | 返回支付链接或二维码 |
| 签名错误 | 返回签名验证失败 |
| 回调签名验证 | 验证通过并处理订单 |
| 重复回调 | 幂等性处理，不重复扣款 |
| 订单查询 | 返回正确的订单信息 |
| 退款申请 | 成功发起退款 |

---

## 12. 常见问题

**Q1: 签名验证失败怎么办？**
- 检查参数是否按 ASCII 排序
- 检查是否过滤了 sign、sign_type、空值
- 检查商户密钥是否正确
- 检查 MD5 结果是否转小写

**Q2: 回调通知收不到？**
- 检查 NotifyURL 是否公网可访问
- 检查是否使用 HTTPS
- 检查防火墙和安全组配置

**Q3: 订单重复支付怎么办？**
- 在回调处理中添加幂等性检查
- 使用数据库唯一索引（OutTradeNo）
- 检查订单状态再处理业务逻辑

---

## 13. 更新记录

| 版本 | 日期 | 修改内容 |
|------|------|----------|
| v1.0 | 2025-11-23 | 初始版本 |
| v1.1 | 2025-11-23 | 调整为独立 SDK 包结构，添加使用示例 |
