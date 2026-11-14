// Package auth 提供口令哈希与会话令牌，不依赖项目内其他包。
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// DefaultPasscode 是首次启动时使用的初始口令。
	DefaultPasscode = "123456"
	// MinPasscodeLen 与 MaxPasscodeLen 限制口令长度（按字符计）。
	MinPasscodeLen = 4
	MaxPasscodeLen = 32
	// PasscodeIterations 是 PBKDF2 的迭代次数。
	PasscodeIterations = 120000
	// SessionTTL 是登录会话的有效期。
	SessionTTL = 7 * 24 * time.Hour
)

// Credentials 是保存在 settings 表中的口令摘要。
type Credentials struct {
	Salt       string
	Hash       string
	Iterations int
}

func derive(passcode, salt string, iterations int) []byte {
	if iterations <= 0 {
		iterations = PasscodeIterations
	}
	h := hmac.New(sha256.New, []byte(passcode))
	h.Write([]byte(salt))
	h.Write([]byte{0, 0, 0, 1})
	u := h.Sum(nil)
	out := append([]byte(nil), u...)
	for i := 1; i < iterations; i++ {
		h.Reset()
		h.Write(u)
		u = h.Sum(nil)
		for j := range out {
			out[j] ^= u[j]
		}
	}
	return out
}

// ValidatePasscode 检查口令长度是否符合要求。
func ValidatePasscode(passcode string) error {
	length := len([]rune(passcode))
	if length < MinPasscodeLen || length > MaxPasscodeLen {
		return fmt.Errorf("口令长度需为 %d-%d 位", MinPasscodeLen, MaxPasscodeLen)
	}
	return nil
}

// HashPasscode 生成带随机盐的口令摘要。
func HashPasscode(passcode string) (Credentials, error) {
	if err := ValidatePasscode(passcode); err != nil {
		return Credentials{}, err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return Credentials{}, fmt.Errorf("生成盐失败: %w", err)
	}
	saltText := hex.EncodeToString(salt)
	sum := derive(passcode, saltText, PasscodeIterations)
	return Credentials{
		Salt:       saltText,
		Hash:       base64.RawStdEncoding.EncodeToString(sum),
		Iterations: PasscodeIterations,
	}, nil
}

// VerifyPasscode 用常量时间比较口令摘要。
func VerifyPasscode(passcode string, credentials Credentials) bool {
	if credentials.Salt == "" || credentials.Hash == "" {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(credentials.Hash)
	if err != nil {
		return false
	}
	actual := derive(passcode, credentials.Salt, credentials.Iterations)
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

// RandomToken 生成 n 字节随机数的 URL 安全字符串。
func RandomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

type sessionPayload struct {
	Exp   int64  `json:"exp"`
	Nonce string `json:"n"`
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// SignSession 生成带过期时间的会话令牌。
func SignSession(secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("会话密钥为空")
	}
	nonce, err := RandomToken(12)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(sessionPayload{
		Exp:   time.Now().Add(ttl).Unix(),
		Nonce: nonce,
	})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(body)
	return encoded + "." + sign(secret, encoded), nil
}

// ValidateSession 校验令牌签名与有效期。
func ValidateSession(secret, token string) bool {
	if secret == "" || token == "" {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	expected := sign(secret, parts[0])
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) != 1 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var payload sessionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false
	}
	return payload.Exp > time.Now().Unix()
}
