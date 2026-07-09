package mygin

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// SignHeaderAppID 表示调用方身份，例如 admin-web、client-web
	SignHeaderAppID = "X-App-Id"

	// SignHeaderTimestamp 表示请求发起时间，毫秒时间戳
	SignHeaderTimestamp = "X-Timestamp"

	// SignHeaderNonce 表示本次请求随机串，用于后续防重放
	SignHeaderNonce = "X-Nonce"

	// SignHeaderBodySHA256 表示原始请求 body 的 SHA256 摘要
	SignHeaderBodySHA256 = "X-Body-SHA256"

	// SignHeaderSignature 表示 HMAC-SHA256 签名结果
	SignHeaderSignature = "X-Signature"

	// DefaultSignMaxSkew 默认允许客户端和服务端时间误差，单位秒
	DefaultSignMaxSkew = 300
)

var (
	// ErrSignMissingHeader 表示缺少签名所需 header
	ErrSignMissingHeader = errors.New("missing signature header")

	// ErrSignInvalidTimestamp 表示时间戳格式错误
	ErrSignInvalidTimestamp = errors.New("invalid signature timestamp")

	// ErrSignExpired 表示请求时间戳已过期
	ErrSignExpired = errors.New("signature timestamp expired")

	// ErrSignBodyHashMismatch 表示 body 摘要不一致，请求体可能被篡改
	ErrSignBodyHashMismatch = errors.New("signature body sha256 mismatch")

	// ErrSignMismatch 表示签名不一致，请求可能被篡改或伪造
	ErrSignMismatch = errors.New("signature mismatch")
)

// SignOption 表示调用方生成签名 header 时需要提供的信息
type SignOption struct {
	// AppID 表示调用方身份，例如 admin-web、client-web
	AppID string

	// Method 表示 HTTP 方法，例如 GET、POST
	Method string

	// Path 表示参与签名的请求路径，必须和服务端实际收到的 path 一致
	Path string

	// RawQuery 表示 URL 原始 query，例如 page=1&limit=20
	RawQuery string

	// Body 表示原始请求 body；GET 没有 body 时传 nil 或空切片
	Body []byte

	// Timestamp 表示毫秒时间戳；如果为 0，函数内部自动使用当前时间
	Timestamp int64

	// Nonce 表示随机串；如果为空，函数内部自动生成
	Nonce string
}

// BuildSignHeaders 根据请求信息生成需要放入 HTTP header 的签名字段
func BuildSignHeaders(key string, opt SignOption) (map[string]string, error) {
	if key == "" {
		return nil, errors.New("sign key is empty")
	}
	if opt.AppID == "" {
		return nil, errors.New("app_id is empty")
	}
	if opt.Method == "" {
		return nil, errors.New("method is empty")
	}
	if opt.Path == "" {
		return nil, errors.New("path is empty")
	}

	timestamp := opt.Timestamp
	if timestamp == 0 {
		timestamp = time.Now().UnixMilli()
	}

	nonce := opt.Nonce
	if nonce == "" {
		nonce = NewSignNonce()
	}

	bodyHash := SHA256Hex(opt.Body)
	canonicalQuery, err := canonicalQuery(opt.RawQuery)
	if err != nil {
		return nil, err
	}

	canonicalString := buildCanonicalString(SignCanonicalOption{
		Method:     opt.Method,
		Path:       opt.Path,
		Query:      canonicalQuery,
		BodySHA256: bodyHash,
		Timestamp:  timestamp,
		Nonce:      nonce,
		AppID:      opt.AppID,
	})
	signature := HMACSHA256Hex(key, canonicalString)

	return map[string]string{
		SignHeaderAppID:      opt.AppID,
		SignHeaderTimestamp:  strconv.FormatInt(timestamp, 10),
		SignHeaderNonce:      nonce,
		SignHeaderBodySHA256: bodyHash,
		SignHeaderSignature:  signature,
	}, nil
}

// VerifySignedRequest 验证 HTTP 请求签名；验证后会恢复 request body，避免影响后续 mygin 绑定参数
func VerifySignedRequest(key string, r *http.Request) error {
	if key == "" {
		return errors.New("sign key is empty")
	}
	if r == nil {
		return errors.New("request is nil")
	}

	appID := r.Header.Get(SignHeaderAppID)
	timestampText := r.Header.Get(SignHeaderTimestamp)
	nonce := r.Header.Get(SignHeaderNonce)
	clientBodyHash := r.Header.Get(SignHeaderBodySHA256)
	clientSignature := r.Header.Get(SignHeaderSignature)

	if appID == "" || timestampText == "" || nonce == "" || clientBodyHash == "" || clientSignature == "" {
		return ErrSignMissingHeader
	}

	timestamp, err := strconv.ParseInt(timestampText, 10, 64)
	if err != nil {
		return ErrSignInvalidTimestamp
	}

	if IsSignTimestampExpired(timestamp, DefaultSignMaxSkew) {
		return ErrSignExpired
	}

	bodyBytes, err := readAndRestoreBody(r)
	if err != nil {
		return err
	}

	serverBodyHash := SHA256Hex(bodyBytes)
	if !SecureEqualString(serverBodyHash, clientBodyHash) {
		return ErrSignBodyHashMismatch
	}

	canonicalQuery, err := canonicalQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}

	canonicalString := buildCanonicalString(SignCanonicalOption{
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      canonicalQuery,
		BodySHA256: serverBodyHash,
		Timestamp:  timestamp,
		Nonce:      nonce,
		AppID:      appID,
	})
	serverSignature := HMACSHA256Hex(key, canonicalString)

	if !SecureEqualString(serverSignature, clientSignature) {
		return ErrSignMismatch
	}

	return nil
}

// SignCanonicalOption 表示构造待签名字符串需要的标准字段
type SignCanonicalOption struct {
	// Method 表示 HTTP 方法
	Method string

	// Path 表示请求路径
	Path string

	// Query 表示已经标准化后的 query 字符串
	Query string

	// BodySHA256 表示请求 body 的 SHA256 摘要
	BodySHA256 string

	// Timestamp 表示毫秒时间戳
	Timestamp int64

	// Nonce 表示请求随机串
	Nonce string

	// AppID 表示调用方身份
	AppID string
}

// BuildCanonicalString 构造待签名字符串；调用方和服务端必须使用完全一致的规则
func buildCanonicalString(opt SignCanonicalOption) string {
	return strings.Join([]string{
		strings.ToUpper(opt.Method),
		opt.Path,
		opt.Query,
		opt.BodySHA256,
		strconv.FormatInt(opt.Timestamp, 10),
		opt.Nonce,
		opt.AppID,
	}, "\n")
}

// CanonicalQuery 将原始 query 标准化，避免参数顺序不同导致签名不一致
func canonicalQuery(rawQuery string) (string, error) {
	if rawQuery == "" {
		return "", nil
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", err
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0)
	for _, key := range keys {
		items := values[key]
		sort.Strings(items)

		escapedKey := url.QueryEscape(key)
		for _, value := range items {
			parts = append(parts, fmt.Sprintf("%s=%s", escapedKey, url.QueryEscape(value)))
		}
	}

	return strings.Join(parts, "&"), nil
}

// SHA256Hex 计算数据的 SHA256，并返回 hex 字符串
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HMACSHA256Hex 使用 key 对 data 做 HMAC-SHA256 签名，并返回 hex 字符串
func HMACSHA256Hex(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// SecureEqualString 使用固定时间比较字符串，避免时序攻击
func SecureEqualString(a string, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}

// IsSignTimestampExpired 判断签名时间戳是否过期
func IsSignTimestampExpired(timestampMilli int64, maxSkewSeconds int64) bool {
	now := time.Now().UnixMilli()
	maxSkewMilli := maxSkewSeconds * 1000

	diff := now - timestampMilli
	if diff < 0 {
		diff = -diff
	}

	return diff > maxSkewMilli
}

// ReadAndRestoreBody 读取 request body，并重新写回，避免后续业务绑定参数时 body 为空
func readAndRestoreBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte{}, nil
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return bodyBytes, nil
}

// NewSignNonce 生成签名用随机 nonce
func NewSignNonce() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(buf)
}
