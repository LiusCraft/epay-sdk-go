package epay

import (
	"strings"
	"time"
)

const (
	// DefaultTimeout 默认超时时间（秒）
	DefaultTimeout = 30

	// DefaultSignType 默认签名类型
	DefaultSignType = "MD5"
)

// Config EPay SDK 配置
type Config struct {
	PID        int    // 商户ID
	Key        string // 商户密钥
	APIBaseURL string // API 基础URL（如: https://pay.example.com）
	Timeout    int    // 请求超时时间（秒，默认: 30）
	Debug      bool   // 是否开启调试模式
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	if c.PID <= 0 {
		return ErrInvalidPID
	}
	if c.Key == "" {
		return ErrInvalidKey
	}
	if c.APIBaseURL == "" {
		return ErrInvalidAPIURL
	}
	return nil
}

// GetTimeout 获取超时时间
func (c *Config) GetTimeout() time.Duration {
	if c.Timeout <= 0 {
		return time.Duration(DefaultTimeout) * time.Second
	}
	return time.Duration(c.Timeout) * time.Second
}

// GetAPIBaseURL 获取 API 基础 URL（去除尾部斜杠）
func (c *Config) GetAPIBaseURL() string {
	return strings.TrimRight(c.APIBaseURL, "/")
}
