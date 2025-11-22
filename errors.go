package epay

import "fmt"

// 错误码常量
const (
	ErrCodeInvalidConfig   = 1001 // 配置错误
	ErrCodeSignFailed      = 1002 // 签名失败
	ErrCodeVerifyFailed    = 1003 // 验证失败
	ErrCodeAPIError        = 1004 // API 错误
	ErrCodeNetworkError    = 1005 // 网络错误
	ErrCodeInvalidResponse = 1006 // 响应格式错误
	ErrCodeInvalidParam    = 1007 // 参数错误
)

// EPayError SDK 错误
type EPayError struct {
	Code    int    // 错误码
	Message string // 错误信息
	Err     error  // 原始错误
}

// Error 实现 error 接口
func (e *EPayError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("epay error [%d]: %s, caused by: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("epay error [%d]: %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *EPayError) Unwrap() error {
	return e.Err
}

// NewError 创建新的 SDK 错误
func NewError(code int, message string) *EPayError {
	return &EPayError{
		Code:    code,
		Message: message,
	}
}

// WrapError 包装错误
func WrapError(code int, message string, err error) *EPayError {
	return &EPayError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 预定义错误
var (
	ErrInvalidPID    = NewError(ErrCodeInvalidConfig, "invalid PID: must be greater than 0")
	ErrInvalidKey    = NewError(ErrCodeInvalidConfig, "invalid Key: must not be empty")
	ErrInvalidAPIURL = NewError(ErrCodeInvalidConfig, "invalid APIBaseURL: must not be empty")

	ErrMissingOutTradeNo = NewError(ErrCodeInvalidParam, "out_trade_no is required")
	ErrMissingNotifyURL  = NewError(ErrCodeInvalidParam, "notify_url is required")
	ErrMissingName       = NewError(ErrCodeInvalidParam, "name is required")
	ErrInvalidMoney      = NewError(ErrCodeInvalidParam, "money must be greater than 0")
	ErrMissingTradeNo    = NewError(ErrCodeInvalidParam, "trade_no or out_trade_no is required")

	ErrSignVerifyFailed = NewError(ErrCodeVerifyFailed, "signature verification failed")
)
