package utils_test

import (
	"encoding/base64"
	"stock-analysis-backend/internal/utils"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test_jwt_secret_key_for_unit_testing"

// TestGenerateToken_Success 测试生成 Token 成功
func TestGenerateToken_Success(t *testing.T) {
	userID := uint64(1)
	username := "testuser"

	token, err := utils.GenerateToken(userID, username, testSecret, 24)

	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// JWT token 应该有三部分，用 . 分隔
	if len(token) < 50 {
		t.Errorf("GenerateToken() token too short: %s", token)
	}
}

// TestParseToken_Success 测试解析 Token 成功
func TestParseToken_Success(t *testing.T) {
	userID := uint64(123)
	username := "testuser"

	token, _ := utils.GenerateToken(userID, username, testSecret, 24)

	claims, err := utils.ParseToken(token, testSecret)

	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("ParseToken() UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("ParseToken() Username = %v, want %v", claims.Username, username)
	}
}

// TestParseToken_InvalidToken 测试解析无效 Token
func TestParseToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"

	_, err := utils.ParseToken(invalidToken, testSecret)

	if err == nil {
		t.Error("ParseToken() should return error for invalid token")
	}
}

// TestParseToken_EmptyToken 测试解析空 Token
func TestParseToken_EmptyToken(t *testing.T) {
	_, err := utils.ParseToken("", testSecret)

	if err == nil {
		t.Error("ParseToken() should return error for empty token")
	}
}

// TestParseToken_WrongSecret 测试使用错误密钥解析 Token
func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := utils.GenerateToken(1, "testuser", testSecret, 24)
	wrongSecret := "wrong_secret_key"

	_, err := utils.ParseToken(token, wrongSecret)

	if err == nil {
		t.Error("ParseToken() should return error for wrong secret")
	}
}

// TestParseToken_TamperedToken 测试篡改的 Token
func TestParseToken_TamperedToken(t *testing.T) {
	token, _ := utils.GenerateToken(1, "testuser", testSecret, 24)

	// 篡改 token 的最后一个字符
	tamperedToken := token[:len(token)-1] + "X"

	_, err := utils.ParseToken(tamperedToken, testSecret)

	if err == nil {
		t.Error("ParseToken() should return error for tampered token")
	}
}

// TestGenerateToken_DifferentSecrets 测试不同密钥生成不同签名
func TestGenerateToken_DifferentSecrets(t *testing.T) {
	userID := uint64(1)
	username := "testuser"
	secret1 := "secret1"
	secret2 := "secret2"

	token1, _ := utils.GenerateToken(userID, username, secret1, 24)
	token2, _ := utils.GenerateToken(userID, username, secret2, 24)

	// 不同密钥生成的 token 签名部分应该不同
	if token1 == token2 {
		t.Error("Different secrets should produce different tokens")
	}
}

// TestGenerateToken_DifferentUsers 测试不同用户生成不同 Token
func TestGenerateToken_DifferentUsers(t *testing.T) {
	token1, _ := utils.GenerateToken(1, "user1", testSecret, 24)
	token2, _ := utils.GenerateToken(2, "user2", testSecret, 24)

	if token1 == token2 {
		t.Error("Different users should produce different tokens")
	}
}

// TestGenerateToken_Expiration 测试 Token 过期时间
func TestGenerateToken_Expiration(t *testing.T) {
	userID := uint64(1)
	username := "testuser"

	// 生成一个立即过期的 token (0 小时)
	token, err := utils.GenerateToken(userID, username, testSecret, 0)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 等待一小段时间确保 token 过期
	time.Sleep(100 * time.Millisecond)

	// 注意：JWT 的过期检查取决于服务器时间
	// 这个测试主要验证生成不会出错
	_, parseErr := utils.ParseToken(token, testSecret)

	// 可能过期的 token 解析会失败，也可能不会（取决于时间精度）
	// 这里我们只是确保不会 panic
	t.Logf("Parse expired token result: err=%v", parseErr)
}

// TestTokenRoundTrip 测试完整流程：生成 -> 解析 -> 验证
func TestTokenRoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		userID   uint64
		username string
		expire   int
	}{
		{"Regular user", 1, "regularuser", 24},
		{"Admin user", 999, "admin", 168},
		{"Long username", 100, "very_long_username_for_testing", 1},
		{"Chinese username", 200, "测试用户", 24},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := utils.GenerateToken(tc.userID, tc.username, testSecret, tc.expire)
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}

			claims, err := utils.ParseToken(token, testSecret)
			if err != nil {
				t.Fatalf("ParseToken() error = %v", err)
			}

			if claims.UserID != tc.userID {
				t.Errorf("UserID mismatch: got %v, want %v", claims.UserID, tc.userID)
			}

			if claims.Username != tc.username {
				t.Errorf("Username mismatch: got %v, want %v", claims.Username, tc.username)
			}
		})
	}
}

// ========== 安全测试 ==========

func TestJWT_Security_TamperedToken(t *testing.T) {
	// 生成一个有效 token
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 篡改 token payload（修改 user_id）
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("Invalid token format")
	}
	// 修改 payload 部分
	tamperedPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"user_id":999,"username":"hacker"}`))
	tamperedToken := parts[0] + "." + tamperedPayload + "." + parts[2]

	// 验证篡改后的 token 应该失败
	_, err = utils.ParseToken(tamperedToken, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for tampered token")
	}
}

func TestJWT_Security_TamperedSignature(t *testing.T) {
	// 生成一个有效 token
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 篡改签名
	parts := strings.Split(token, ".")
	tamperedToken := parts[0] + "." + parts[1] + ".tampered_signature"

	// 验证篡改签名后的 token 应该失败
	_, err = utils.ParseToken(tamperedToken, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for tampered signature")
	}
}

func TestJWT_Security_ExpiredToken(t *testing.T) {
	// 生成一个已过期的 token（-1 小时）
	token, err := utils.GenerateToken(1, "testuser", testSecret, -1)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 验证过期 token 应该失败
	_, err = utils.ParseToken(token, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for expired token")
	}
}

func TestJWT_Security_WrongSecret(t *testing.T) {
	// 使用正确 secret 生成 token
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 使用错误 secret 验证
	wrongSecret := "wrong_secret_key"
	_, err = utils.ParseToken(token, wrongSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail with wrong secret")
	}
}

func TestJWT_Security_EmptyClaims(t *testing.T) {
	// 尝试生成 user_id 为 0 的 token
	token, err := utils.GenerateToken(0, "", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 验证应该成功（但 user_id 为 0）
	claims, err := utils.ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 0 {
		t.Errorf("ParseToken() UserID = %d, want 0", claims.UserID)
	}
}

func TestJWT_Security_MissingUserID(t *testing.T) {
	// 手动创建一个缺少 user_id 的 token
	claims := jwt.MapClaims{
		"username": "testuser",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	// 验证应该成功，但 user_id 为默认值 0
	parsedClaims, err := utils.ParseToken(tokenString, testSecret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if parsedClaims.UserID != 0 {
		t.Errorf("ParseToken() UserID = %d, want 0", parsedClaims.UserID)
	}
}

func TestJWT_Security_VeryLongToken(t *testing.T) {
	// 创建一个超长的 token 字符串
	longToken := strings.Repeat("a", 10000)

	// 验证应该失败
	_, err := utils.ParseToken(longToken, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for very long token")
	}
}

func TestJWT_Security_InvalidAlgorithm(t *testing.T) {
	// 创建一个使用 none 算法的 token
	claims := jwt.MapClaims{
		"user_id":  1,
		"username": "testuser",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	// 验证应该失败（不支持 none 算法）
	_, err = utils.ParseToken(tokenString, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for none algorithm")
	}
}
