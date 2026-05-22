package utils_test

import (
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
