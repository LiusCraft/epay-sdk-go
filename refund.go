package epay

import "fmt"

// API 接口路径
const (
	APIPathRefund = "/api.php" // 退款接口
)

// Refund 提交订单退款
func (c *Client) Refund(req *RefundRequest) (*RefundResponse, error) {
	// 验证参数
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 构建请求参数
	params := c.buildBaseParams()
	params["act"] = "refund"
	params["money"] = fmt.Sprintf("%.2f", req.Money)

	// 优先使用商户订单号
	if req.OutTradeNo != "" {
		params["out_trade_no"] = req.OutTradeNo
	} else if req.TradeNo != "" {
		params["trade_no"] = req.TradeNo
	}

	// 发送请求
	body, err := c.doGet(APIPathRefund, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	resp, err := parseJSONResponse[RefundResponse](body)
	if err != nil {
		return nil, err
	}

	// 检查业务错误
	if resp.Code != 1 {
		return nil, NewError(ErrCodeAPIError, resp.Msg)
	}

	return resp, nil
}

// RefundByOutTradeNo 通过商户订单号退款（便捷方法）
func (c *Client) RefundByOutTradeNo(outTradeNo string, money float64) (*RefundResponse, error) {
	return c.Refund(&RefundRequest{
		OutTradeNo: outTradeNo,
		Money:      money,
	})
}

// RefundByTradeNo 通过 EPay 订单号退款（便捷方法）
func (c *Client) RefundByTradeNo(tradeNo string, money float64) (*RefundResponse, error) {
	return c.Refund(&RefundRequest{
		TradeNo: tradeNo,
		Money:   money,
	})
}

// IsRefundSuccess 检查退款是否成功
func IsRefundSuccess(resp *RefundResponse) bool {
	return resp != nil && resp.Code == 1
}
