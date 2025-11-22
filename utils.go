package epay

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// FilterEmptyParams 过滤空值和特殊参数（sign、sign_type）
func FilterEmptyParams(params map[string]string) map[string]string {
	filtered := make(map[string]string)
	for k, v := range params {
		// 跳过空值
		if v == "" {
			continue
		}
		// 跳过签名相关参数
		if k == "sign" || k == "sign_type" {
			continue
		}
		filtered[k] = v
	}
	return filtered
}

// SortAndBuildQuery 按 ASCII 排序参数并构建查询字符串
func SortAndBuildQuery(params map[string]string) string {
	// 获取所有 key 并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建查询字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}

	return strings.Join(parts, "&")
}

// BuildURLQuery 构建 URL 编码的查询字符串
func BuildURLQuery(params map[string]string) string {
	// 获取所有 key 并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建 URL 编码的查询字符串
	values := url.Values{}
	for _, k := range keys {
		values.Set(k, params[k])
	}

	return values.Encode()
}

// ParseNotifyParams 从 HTTP 请求中解析回调参数
// 支持 GET 和 POST 请求
func ParseNotifyParams(r *http.Request) map[string]string {
	params := make(map[string]string)

	// 先尝试解析 URL 查询参数
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 如果是 POST 请求，解析表单数据
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			for k, v := range r.PostForm {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		}
	}

	return params
}

// MapToURLValues 将 map 转换为 url.Values
func MapToURLValues(params map[string]string) url.Values {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values
}
