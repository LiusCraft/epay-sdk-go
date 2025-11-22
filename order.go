package epay

import "strconv"

// API 接口路径
const (
	APIPathQuery  = "/api.php" // 订单查询接口
	APIPathOrders = "/api.php" // 批量订单查询接口
)

// QueryOrder 查询单个订单
func (c *Client) QueryOrder(req *OrderQueryRequest) (*OrderDetail, error) {
	// 验证参数
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 构建请求参数
	params := c.buildBaseParams()
	params["act"] = "order"

	// 优先使用商户订单号
	if req.OutTradeNo != "" {
		params["out_trade_no"] = req.OutTradeNo
	} else if req.TradeNo != "" {
		params["trade_no"] = req.TradeNo
	}

	// 发送请求
	body, err := c.doGet(APIPathQuery, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	resp, err := parseJSONResponse[OrderDetail](body)
	if err != nil {
		return nil, err
	}

	// 检查业务错误
	if resp.Code != 1 {
		return nil, NewError(ErrCodeAPIError, resp.Msg)
	}

	return resp, nil
}

// QueryOrders 批量查询订单
func (c *Client) QueryOrders(limit, page int) (*OrderListResponse, error) {
	// 参数校验
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}

	// 构建请求参数
	params := c.buildBaseParams()
	params["act"] = "orders"
	params["limit"] = strconv.Itoa(limit)
	params["page"] = strconv.Itoa(page)

	// 发送请求
	body, err := c.doGet(APIPathOrders, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	resp, err := parseJSONResponse[OrderListResponse](body)
	if err != nil {
		return nil, err
	}

	// 检查业务错误
	if resp.Code != 1 {
		return nil, NewError(ErrCodeAPIError, resp.Msg)
	}

	return resp, nil
}

// IsOrderPaid 检查订单是否已支付
func IsOrderPaid(order *OrderDetail) bool {
	return order != nil && order.Status == OrderStatusPaid
}

// IsOrderUnpaid 检查订单是否未支付
func IsOrderUnpaid(order *OrderDetail) bool {
	return order != nil && order.Status == OrderStatusUnpaid
}
