package epay

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Client EPay SDK 客户端
type Client struct {
	config     *Config
	httpClient *http.Client
	signer     *Signer
}

// NewClient 创建 EPay 客户端
func NewClient(config *Config) (*Client, error) {
	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 创建 HTTP 客户端
	httpClient := &http.Client{
		Timeout: config.GetTimeout(),
	}

	// 创建签名器
	signer := NewSigner(config.Key)

	return &Client{
		config:     config,
		httpClient: httpClient,
		signer:     signer,
	}, nil
}

// GetConfig 获取配置（只读）
func (c *Client) GetConfig() Config {
	return *c.config
}

// buildBaseParams 构建基础请求参数
func (c *Client) buildBaseParams() map[string]string {
	return map[string]string{
		"pid": strconv.Itoa(c.config.PID),
	}
}

// doGet 执行 GET 请求
func (c *Client) doGet(endpoint string, params map[string]string) ([]byte, error) {
	// 构建 URL
	reqURL := c.config.GetAPIBaseURL() + endpoint

	// 添加签名
	signedParams := c.signer.SignWithParams(params)

	// 构建查询字符串
	queryString := BuildURLQuery(signedParams)
	fullURL := reqURL + "?" + queryString

	if c.config.Debug {
		log.Printf("[EPay SDK] GET %s", fullURL)
	}

	// 发送请求
	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, WrapError(ErrCodeNetworkError, "HTTP request failed", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, WrapError(ErrCodeNetworkError, "read response failed", err)
	}

	if c.config.Debug {
		log.Printf("[EPay SDK] Response: %s", string(body))
	}

	return body, nil
}

// doPost 执行 POST 请求
func (c *Client) doPost(endpoint string, params map[string]string) ([]byte, error) {
	// 构建 URL
	reqURL := c.config.GetAPIBaseURL() + endpoint

	// 添加签名
	signedParams := c.signer.SignWithParams(params)

	// 构建表单数据
	formData := url.Values{}
	for k, v := range signedParams {
		formData.Set(k, v)
	}

	if c.config.Debug {
		log.Printf("[EPay SDK] POST %s, params: %v", reqURL, signedParams)
	}

	// 发送请求
	resp, err := c.httpClient.Post(reqURL, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, WrapError(ErrCodeNetworkError, "HTTP request failed", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, WrapError(ErrCodeNetworkError, "read response failed", err)
	}

	if c.config.Debug {
		log.Printf("[EPay SDK] Response: %s", string(body))
	}

	return body, nil
}

// parseJSONResponse 解析 JSON 响应
func parseJSONResponse[T any](body []byte) (*T, error) {
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, WrapError(ErrCodeInvalidResponse, "parse JSON response failed", err)
	}
	return &result, nil
}

// VerifyNotify 验证支付回调通知
func (c *Client) VerifyNotify(params map[string]string) (*NotifyData, error) {
	// 获取签名
	sign := params["sign"]
	if sign == "" {
		return nil, NewError(ErrCodeVerifyFailed, "missing sign parameter")
	}

	// 验证签名
	if !c.signer.Verify(params, sign) {
		return nil, ErrSignVerifyFailed
	}

	// 解析 PID
	pid, _ := strconv.Atoi(params["pid"])

	// 构建通知数据
	notifyData := &NotifyData{
		PID:         pid,
		TradeNo:     params["trade_no"],
		OutTradeNo:  params["out_trade_no"],
		Type:        params["type"],
		Name:        params["name"],
		Money:       params["money"],
		TradeStatus: params["trade_status"],
		Param:       params["param"],
		Sign:        sign,
		SignType:    params["sign_type"],
	}

	return notifyData, nil
}

// Sign 对参数进行签名（暴露给外部使用）
func (c *Client) Sign(params map[string]string) string {
	return c.signer.Sign(params)
}

// Verify 验证签名（暴露给外部使用）
func (c *Client) Verify(params map[string]string, sign string) bool {
	return c.signer.Verify(params, sign)
}
