package utils_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"stock-analysis-backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test_jwt_secret_key_for_unit_testing"

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

	if len(token) < 50 {
		t.Errorf("GenerateToken() token too short: %s", token)
	}
}

func TestGenerateTokenSetsJTI(t *testing.T) {
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := utils.ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.ID == "" {
		t.Fatal("claims.ID = empty, want non-empty jti")
	}
}

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

func TestParseToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"

	_, err := utils.ParseToken(invalidToken, testSecret)

	if err == nil {
		t.Error("ParseToken() should return error for invalid token")
	}
}

func TestParseToken_EmptyToken(t *testing.T) {
	_, err := utils.ParseToken("", testSecret)

	if err == nil {
		t.Error("ParseToken() should return error for empty token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := utils.GenerateToken(1, "testuser", testSecret, 24)
	wrongSecret := "wrong_secret_key"

	_, err := utils.ParseToken(token, wrongSecret)

	if err == nil {
		t.Error("ParseToken() should return error for wrong secret")
	}
}

func TestParseToken_TamperedToken(t *testing.T) {
	token, _ := utils.GenerateToken(1, "testuser", testSecret, 24)
	tamperedToken := token[:len(token)-1] + "X"

	_, err := utils.ParseToken(tamperedToken, testSecret)

	if err == nil {
		t.Error("ParseToken() should return error for tampered token")
	}
}

func TestGenerateToken_DifferentSecrets(t *testing.T) {
	userID := uint64(1)
	username := "testuser"
	secret1 := "secret1"
	secret2 := "secret2"

	token1, _ := utils.GenerateToken(userID, username, secret1, 24)
	token2, _ := utils.GenerateToken(userID, username, secret2, 24)

	if token1 == token2 {
		t.Error("Different secrets should produce different tokens")
	}
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	token1, _ := utils.GenerateToken(1, "user1", testSecret, 24)
	token2, _ := utils.GenerateToken(2, "user2", testSecret, 24)

	if token1 == token2 {
		t.Error("Different users should produce different tokens")
	}
}

func TestGenerateToken_Expiration(t *testing.T) {
	userID := uint64(1)
	username := "testuser"

	token, err := utils.GenerateToken(userID, username, testSecret, 0)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	_, parseErr := utils.ParseToken(token, testSecret)
	t.Logf("Parse expired token result: err=%v", parseErr)
}

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
			if claims.ID == "" {
				t.Fatal("claims.ID = empty, want non-empty jti")
			}
		})
	}
}

func TestParseLegacyTokenWithoutJTI(t *testing.T) {
	legacyClaims := utils.Claims{
		UserID:   1,
		Username: "legacy-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	legacyToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyClaims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	claims, err := utils.ParseToken(legacyToken, testSecret)
	if err != nil {
		t.Fatalf("ParseToken() legacy token error = %v", err)
	}
	if claims.ID != "" {
		t.Fatalf("legacy claims.ID = %q, want empty", claims.ID)
	}
}

func TestJWT_Security_TamperedToken(t *testing.T) {
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("Invalid token format")
	}
	tamperedPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"user_id":999,"username":"hacker"}`))
	tamperedToken := parts[0] + "." + tamperedPayload + "." + parts[2]

	_, err = utils.ParseToken(tamperedToken, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for tampered token")
	}
}

func TestJWT_Security_TamperedSignature(t *testing.T) {
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	tamperedToken := parts[0] + "." + parts[1] + ".tampered_signature"

	_, err = utils.ParseToken(tamperedToken, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for tampered signature")
	}
}

func TestJWT_Security_ExpiredToken(t *testing.T) {
	token, err := utils.GenerateToken(1, "testuser", testSecret, -1)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	_, err = utils.ParseToken(token, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for expired token")
	}
}

func TestJWT_Security_WrongSecret(t *testing.T) {
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	wrongSecret := "wrong_secret_key"
	_, err = utils.ParseToken(token, wrongSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail with wrong secret")
	}
}

func TestJWT_Security_EmptyClaims(t *testing.T) {
	token, err := utils.GenerateToken(0, "", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := utils.ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 0 {
		t.Errorf("ParseToken() UserID = %d, want 0", claims.UserID)
	}
}

func TestJWT_Security_MissingUserID(t *testing.T) {
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

	parsedClaims, err := utils.ParseToken(tokenString, testSecret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if parsedClaims.UserID != 0 {
		t.Errorf("ParseToken() UserID = %d, want 0", parsedClaims.UserID)
	}
}

func TestJWT_Security_VeryLongToken(t *testing.T) {
	longToken := strings.Repeat("a", 10000)

	_, err := utils.ParseToken(longToken, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for very long token")
	}
}

func TestJWT_Security_InvalidAlgorithm(t *testing.T) {
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

	_, err = utils.ParseToken(tokenString, testSecret)
	if err == nil {
		t.Fatalf("ParseToken() should fail for none algorithm")
	}
}

func BenchmarkJWT_GenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := utils.GenerateToken(1, "testuser", testSecret, 24)
		if err != nil {
			b.Fatalf("GenerateToken() error = %v", err)
		}
	}
}

func BenchmarkJWT_ParseToken(b *testing.B) {
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		b.Fatalf("GenerateToken() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := utils.ParseToken(token, testSecret)
		if err != nil {
			b.Fatalf("ParseToken() error = %v", err)
		}
	}
}

func BenchmarkJWT_ValidateToken(b *testing.B) {
	token, err := utils.GenerateToken(1, "testuser", testSecret, 24)
	if err != nil {
		b.Fatalf("GenerateToken() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := utils.ParseToken(token, testSecret)
		if err != nil {
			b.Fatalf("ParseToken() error = %v", err)
		}
	}
}
