package epay

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

// Validate 验证支付请求参数
func (r *PaymentRequest) Validate() error {
	if r.OutTradeNo == "" {
		return ErrMissingOutTradeNo
	}
	if r.NotifyURL == "" {
		return ErrMissingNotifyURL
	}
	if r.Name == "" {
		return ErrMissingName
	}
	if r.Money <= 0 {
		return ErrInvalidMoney
	}
	return nil
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

// Validate 验证表单支付请求参数
func (r *FormPaymentRequest) Validate() error {
	if r.OutTradeNo == "" {
		return ErrMissingOutTradeNo
	}
	if r.NotifyURL == "" {
		return ErrMissingNotifyURL
	}
	if r.Name == "" {
		return ErrMissingName
	}
	if r.Money <= 0 {
		return ErrInvalidMoney
	}
	return nil
}

// PaymentResponse API 接口支付响应
type PaymentResponse struct {
	Code      int    `json:"code"`      // 1=成功，其他=失败
	Msg       string `json:"msg"`       // 错误信息
	TradeNo   string `json:"trade_no"`  // 支付订单号
	PayURL    string `json:"payurl"`    // 支付跳转URL
	QRCode    string `json:"qrcode"`    // 二维码链接
	URLScheme string `json:"urlscheme"` // 小程序跳转URL
}

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

// OrderQueryRequest 订单查询请求
type OrderQueryRequest struct {
	TradeNo    string // EPay订单号（二选一）
	OutTradeNo string // 商户订单号（二选一）
}

// Validate 验证订单查询请求参数
func (r *OrderQueryRequest) Validate() error {
	if r.TradeNo == "" && r.OutTradeNo == "" {
		return ErrMissingTradeNo
	}
	return nil
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

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Code   int           `json:"code"`
	Msg    string        `json:"msg"`
	Count  int           `json:"count"`
	Orders []OrderDetail `json:"orders"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	TradeNo    string  // EPay订单号（二选一）
	OutTradeNo string  // 商户订单号（二选一）
	Money      float64 // 退款金额
}

// Validate 验证退款请求参数
func (r *RefundRequest) Validate() error {
	if r.TradeNo == "" && r.OutTradeNo == "" {
		return ErrMissingTradeNo
	}
	if r.Money <= 0 {
		return ErrInvalidMoney
	}
	return nil
}

// RefundResponse 退款响应
type RefundResponse struct {
	Code int    `json:"code"` // 1=成功
	Msg  string `json:"msg"`
}

// 支付状态常量
const (
	TradeStatusSuccess = "TRADE_SUCCESS" // 支付成功
)

// 订单状态常量
const (
	OrderStatusUnpaid = 0 // 未支付
	OrderStatusPaid   = 1 // 已支付
)

// 支付方式常量
const (
	PayTypeAlipay = "alipay" // 支付宝
	PayTypeWxpay  = "wxpay"  // 微信支付
	PayTypeQQpay  = "qqpay"  // QQ钱包
)

// 设备类型常量
const (
	DevicePC     = "pc"     // PC端
	DeviceMobile = "mobile" // 移动端
	DeviceWechat = "wechat" // 微信内
	DeviceAlipay = "alipay" // 支付宝内
)
