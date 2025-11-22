package epay

import "time"

// ClientBuilder 客户端构建器，支持链式调用
type ClientBuilder struct {
	config *Config
}

// New 创建一个新的客户端构建器（链式 API）
// 示例:
//
//	client := epay.New(1001, "your-key", "https://pay.example.com").
//	    WithTimeout(30).
//	    WithDebug(true).
//	    Build()
func New(pid int, key string, apiBaseURL string) *ClientBuilder {
	return &ClientBuilder{
		config: &Config{
			PID:        pid,
			Key:        key,
			APIBaseURL: apiBaseURL,
			Timeout:    30, // 默认 30 秒
			Debug:      false,
		},
	}
}

// NewQuick 快速创建客户端（一行代码）
// 使用默认配置：超时 30 秒，关闭调试
// 示例:
//
//	client := epay.NewQuick(1001, "your-key", "https://pay.example.com")
func NewQuick(pid int, key string, apiBaseURL string) *Client {
	client, err := New(pid, key, apiBaseURL).Build()
	if err != nil {
		// 快速创建模式：配置错误直接 panic
		panic(err)
	}
	return client
}

// WithTimeout 设置超时时间（秒）
func (b *ClientBuilder) WithTimeout(seconds int) *ClientBuilder {
	b.config.Timeout = seconds
	return b
}

// WithDebug 设置调试模式
func (b *ClientBuilder) WithDebug(debug bool) *ClientBuilder {
	b.config.Debug = debug
	return b
}

// WithHTTPTimeout 设置 HTTP 超时时间（time.Duration）
func (b *ClientBuilder) WithHTTPTimeout(timeout time.Duration) *ClientBuilder {
	b.config.Timeout = int(timeout.Seconds())
	return b
}

// Build 构建客户端
func (b *ClientBuilder) Build() (*Client, error) {
	return NewClient(b.config)
}

// MustBuild 构建客户端，出错则 panic
func (b *ClientBuilder) MustBuild() *Client {
	client, err := b.Build()
	if err != nil {
		panic(err)
	}
	return client
}
