package epay

import (
	"fmt"
	"html"
	"strconv"
)

// API 接口路径
const (
	APIPathMapi   = "/mapi.php"   // API 接口支付
	APIPathSubmit = "/submit.php" // 页面跳转支付
)

// CreatePayment 创建 API 接口支付
// 返回支付链接、二维码等信息
func (c *Client) CreatePayment(req *PaymentRequest) (*PaymentResponse, error) {
	// 验证参数
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 构建请求参数
	params := c.buildBaseParams()
	params["type"] = req.Type
	params["out_trade_no"] = req.OutTradeNo
	params["notify_url"] = req.NotifyURL
	params["name"] = req.Name
	params["money"] = fmt.Sprintf("%.2f", req.Money)

	// 可选参数
	if req.ReturnURL != "" {
		params["return_url"] = req.ReturnURL
	}
	if req.ClientIP != "" {
		params["clientip"] = req.ClientIP
	}
	if req.Device != "" {
		params["device"] = req.Device
	}
	if req.Param != "" {
		params["param"] = req.Param
	}

	// 发送请求
	body, err := c.doGet(APIPathMapi, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	resp, err := parseJSONResponse[PaymentResponse](body)
	if err != nil {
		return nil, err
	}

	// 检查业务错误
	if resp.Code != 1 {
		return nil, NewError(ErrCodeAPIError, resp.Msg)
	}

	return resp, nil
}

// BuildFormPaymentURL 构建页面跳转支付 URL
// 返回完整的支付跳转 URL
func (c *Client) BuildFormPaymentURL(req *FormPaymentRequest) (string, error) {
	// 验证参数
	if err := req.Validate(); err != nil {
		return "", err
	}

	// 构建请求参数
	params := c.buildBaseParams()
	params["out_trade_no"] = req.OutTradeNo
	params["notify_url"] = req.NotifyURL
	params["return_url"] = req.ReturnURL
	params["name"] = req.Name
	params["money"] = fmt.Sprintf("%.2f", req.Money)

	// 可选参数
	if req.Type != "" {
		params["type"] = req.Type
	}
	if req.Param != "" {
		params["param"] = req.Param
	}

	// 添加签名
	signedParams := c.signer.SignWithParams(params)

	// 构建 URL
	queryString := BuildURLQuery(signedParams)
	payURL := c.config.GetAPIBaseURL() + APIPathSubmit + "?" + queryString

	return payURL, nil
}

// BuildFormPayment 构建页面跳转支付 HTML 表单
// 返回自动提交的 HTML 表单，可直接输出到浏览器
func (c *Client) BuildFormPayment(req *FormPaymentRequest) (string, error) {
	// 验证参数
	if err := req.Validate(); err != nil {
		return "", err
	}

	// 构建请求参数
	params := c.buildBaseParams()
	params["out_trade_no"] = req.OutTradeNo
	params["notify_url"] = req.NotifyURL
	params["return_url"] = req.ReturnURL
	params["name"] = req.Name
	params["money"] = fmt.Sprintf("%.2f", req.Money)

	// 可选参数
	if req.Type != "" {
		params["type"] = req.Type
	}
	if req.Param != "" {
		params["param"] = req.Param
	}

	// 添加签名
	signedParams := c.signer.SignWithParams(params)

	// 构建 HTML 表单
	formHTML := buildAutoSubmitForm(c.config.GetAPIBaseURL()+APIPathSubmit, signedParams)

	return formHTML, nil
}

// buildAutoSubmitForm 构建自动提交的 HTML 表单
func buildAutoSubmitForm(action string, params map[string]string) string {
	// 构建表单字段
	var fields string
	for k, v := range params {
		fields += fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`,
			html.EscapeString(k), html.EscapeString(v))
	}

	// 构建完整的 HTML
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>正在跳转到支付页面...</title>
</head>
<body>
    <form id="payForm" method="POST" action="%s">
        %s
    </form>
    <script>document.getElementById('payForm').submit();</script>
</body>
</html>`

	return fmt.Sprintf(htmlTemplate, html.EscapeString(action), fields)
}

// GetPayTypes 获取支持的支付方式（静态方法）
func GetPayTypes() map[string]string {
	return map[string]string{
		PayTypeAlipay: "支付宝",
		PayTypeWxpay:  "微信支付",
		PayTypeQQpay:  "QQ钱包",
	}
}

// FormatMoney 格式化金额为字符串（保留两位小数）
func FormatMoney(money float64) string {
	return fmt.Sprintf("%.2f", money)
}

// ParseMoney 解析金额字符串为 float64
func ParseMoney(moneyStr string) (float64, error) {
	return strconv.ParseFloat(moneyStr, 64)
}
