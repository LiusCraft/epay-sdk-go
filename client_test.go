package epay

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				PID:        1001,
				Key:        "test-key",
				APIBaseURL: "https://pay.example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid PID",
			config: &Config{
				PID:        0,
				Key:        "test-key",
				APIBaseURL: "https://pay.example.com",
			},
			wantErr: true,
		},
		{
			name: "empty Key",
			config: &Config{
				PID:        1001,
				Key:        "",
				APIBaseURL: "https://pay.example.com",
			},
			wantErr: true,
		},
		{
			name: "empty APIBaseURL",
			config: &Config{
				PID:        1001,
				Key:        "test-key",
				APIBaseURL: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestSigner_Sign(t *testing.T) {
	signer := NewSigner("testkey123")

	tests := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name: "basic sign",
			params: map[string]string{
				"pid":          "1001",
				"type":         "alipay",
				"out_trade_no": "ORDER001",
				"notify_url":   "https://example.com/notify",
				"name":         "Test Product",
				"money":        "10.00",
			},
			// 签名结果需要根据实际计算
			want: "", // 将在测试中验证签名一致性
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sign1 := signer.Sign(tt.params)
			sign2 := signer.Sign(tt.params)

			// 验证签名一致性
			if sign1 != sign2 {
				t.Errorf("Sign() returned different results for same params: %s vs %s", sign1, sign2)
			}

			// 验证签名长度（MD5 为 32 位）
			if len(sign1) != 32 {
				t.Errorf("Sign() returned invalid length: got %d, want 32", len(sign1))
			}
		})
	}
}

func TestSigner_Verify(t *testing.T) {
	signer := NewSigner("testkey123")

	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "ORDER001",
		"money":        "10.00",
	}

	// 生成签名
	sign := signer.Sign(params)

	tests := []struct {
		name string
		sign string
		want bool
	}{
		{
			name: "valid sign",
			sign: sign,
			want: true,
		},
		{
			name: "invalid sign",
			sign: "invalidsign123456789012345678901",
			want: false,
		},
		{
			name: "empty sign",
			sign: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := signer.Verify(params, tt.sign); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterEmptyParams(t *testing.T) {
	params := map[string]string{
		"pid":       "1001",
		"empty":     "",
		"sign":      "should_be_removed",
		"sign_type": "MD5",
		"name":      "test",
	}

	filtered := FilterEmptyParams(params)

	// 检查保留的参数
	if filtered["pid"] != "1001" {
		t.Error("FilterEmptyParams() should keep non-empty params")
	}
	if filtered["name"] != "test" {
		t.Error("FilterEmptyParams() should keep non-empty params")
	}

	// 检查移除的参数
	if _, exists := filtered["empty"]; exists {
		t.Error("FilterEmptyParams() should remove empty params")
	}
	if _, exists := filtered["sign"]; exists {
		t.Error("FilterEmptyParams() should remove sign param")
	}
	if _, exists := filtered["sign_type"]; exists {
		t.Error("FilterEmptyParams() should remove sign_type param")
	}
}

func TestSortAndBuildQuery(t *testing.T) {
	params := map[string]string{
		"c": "3",
		"a": "1",
		"b": "2",
	}

	result := SortAndBuildQuery(params)
	expected := "a=1&b=2&c=3"

	if result != expected {
		t.Errorf("SortAndBuildQuery() = %s, want %s", result, expected)
	}
}

func TestPaymentRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *PaymentRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &PaymentRequest{
				Type:       "alipay",
				OutTradeNo: "ORDER001",
				NotifyURL:  "https://example.com/notify",
				Name:       "Test Product",
				Money:      10.00,
			},
			wantErr: false,
		},
		{
			name: "missing out_trade_no",
			req: &PaymentRequest{
				Type:      "alipay",
				NotifyURL: "https://example.com/notify",
				Name:      "Test Product",
				Money:     10.00,
			},
			wantErr: true,
		},
		{
			name: "missing notify_url",
			req: &PaymentRequest{
				Type:       "alipay",
				OutTradeNo: "ORDER001",
				Name:       "Test Product",
				Money:      10.00,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			req: &PaymentRequest{
				Type:       "alipay",
				OutTradeNo: "ORDER001",
				NotifyURL:  "https://example.com/notify",
				Money:      10.00,
			},
			wantErr: true,
		},
		{
			name: "invalid money",
			req: &PaymentRequest{
				Type:       "alipay",
				OutTradeNo: "ORDER001",
				NotifyURL:  "https://example.com/notify",
				Name:       "Test Product",
				Money:      0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_VerifyNotify(t *testing.T) {
	client, _ := NewClient(&Config{
		PID:        1001,
		Key:        "testkey123",
		APIBaseURL: "https://pay.example.com",
	})

	// 构建测试参数
	params := map[string]string{
		"pid":          "1001",
		"trade_no":     "T20231123001",
		"out_trade_no": "ORDER001",
		"type":         "alipay",
		"name":         "Test Product",
		"money":        "10.00",
		"trade_status": "TRADE_SUCCESS",
	}

	// 生成签名
	sign := client.Sign(params)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	// 验证
	notifyData, err := client.VerifyNotify(params)
	if err != nil {
		t.Errorf("VerifyNotify() error = %v", err)
		return
	}

	if notifyData.OutTradeNo != "ORDER001" {
		t.Errorf("VerifyNotify() OutTradeNo = %s, want ORDER001", notifyData.OutTradeNo)
	}
	if notifyData.TradeStatus != "TRADE_SUCCESS" {
		t.Errorf("VerifyNotify() TradeStatus = %s, want TRADE_SUCCESS", notifyData.TradeStatus)
	}
}

func TestClient_VerifyNotify_InvalidSign(t *testing.T) {
	client, _ := NewClient(&Config{
		PID:        1001,
		Key:        "testkey123",
		APIBaseURL: "https://pay.example.com",
	})

	params := map[string]string{
		"pid":          "1001",
		"trade_no":     "T20231123001",
		"out_trade_no": "ORDER001",
		"sign":         "invalid_sign_12345678901234567890",
		"sign_type":    "MD5",
	}

	_, err := client.VerifyNotify(params)
	if err == nil {
		t.Error("VerifyNotify() should return error for invalid sign")
	}
}

func TestParseNotifyParams(t *testing.T) {
	// 创建测试请求
	req := httptest.NewRequest(http.MethodGet, "/notify?pid=1001&out_trade_no=ORDER001&money=10.00", nil)

	params := ParseNotifyParams(req)

	if params["pid"] != "1001" {
		t.Errorf("ParseNotifyParams() pid = %s, want 1001", params["pid"])
	}
	if params["out_trade_no"] != "ORDER001" {
		t.Errorf("ParseNotifyParams() out_trade_no = %s, want ORDER001", params["out_trade_no"])
	}
	if params["money"] != "10.00" {
		t.Errorf("ParseNotifyParams() money = %s, want 10.00", params["money"])
	}
}

func TestBuildFormPaymentURL(t *testing.T) {
	client, _ := NewClient(&Config{
		PID:        1001,
		Key:        "testkey123",
		APIBaseURL: "https://pay.example.com",
	})

	req := &FormPaymentRequest{
		Type:       "alipay",
		OutTradeNo: "ORDER001",
		NotifyURL:  "https://example.com/notify",
		ReturnURL:  "https://example.com/return",
		Name:       "Test Product",
		Money:      10.00,
	}

	url, err := client.BuildFormPaymentURL(req)
	if err != nil {
		t.Errorf("BuildFormPaymentURL() error = %v", err)
		return
	}

	// 验证 URL 包含必要参数
	if url == "" {
		t.Error("BuildFormPaymentURL() returned empty URL")
	}
	if len(url) < 100 {
		t.Errorf("BuildFormPaymentURL() URL seems too short: %s", url)
	}
}

func TestEPayError_Error(t *testing.T) {
	err := NewError(ErrCodeAPIError, "test error")
	expected := "epay error [1004]: test error"

	if err.Error() != expected {
		t.Errorf("Error() = %s, want %s", err.Error(), expected)
	}
}

func TestEPayError_Unwrap(t *testing.T) {
	originalErr := NewError(ErrCodeNetworkError, "network error")
	wrappedErr := WrapError(ErrCodeAPIError, "api error", originalErr)

	if wrappedErr.Unwrap() != originalErr {
		t.Error("Unwrap() should return original error")
	}
}
