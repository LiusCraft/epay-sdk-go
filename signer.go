package epay

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// Signer MD5 签名器
type Signer struct {
	key string // 商户密钥
}

// NewSigner 创建签名器
func NewSigner(key string) *Signer {
	return &Signer{key: key}
}

// Sign 生成 MD5 签名
// 签名流程：
// 1. 过滤参数：去除 sign、sign_type、空值
// 2. 按 ASCII 排序参数名
// 3. 拼接为 a=b&c=d 格式
// 4. 拼接商户密钥 KEY
// 5. MD5 加密并转小写
func (s *Signer) Sign(params map[string]string) string {
	// 1. 过滤空值和特殊参数
	filtered := FilterEmptyParams(params)

	// 2 & 3. 排序并构建查询字符串
	queryString := SortAndBuildQuery(filtered)

	// 4. 拼接商户密钥
	signString := queryString + s.key

	// 5. MD5 加密并转小写
	hash := md5.Sum([]byte(signString))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

// Verify 验证签名是否正确
func (s *Signer) Verify(params map[string]string, receivedSign string) bool {
	// 计算签名
	expectedSign := s.Sign(params)

	// 比较签名（忽略大小写）
	return strings.EqualFold(expectedSign, receivedSign)
}

// SignWithParams 对参数进行签名并返回包含签名的参数
func (s *Signer) SignWithParams(params map[string]string) map[string]string {
	// 复制参数
	signedParams := make(map[string]string)
	for k, v := range params {
		signedParams[k] = v
	}

	// 计算签名
	sign := s.Sign(params)

	// 添加签名
	signedParams["sign"] = sign
	signedParams["sign_type"] = DefaultSignType

	return signedParams
}
