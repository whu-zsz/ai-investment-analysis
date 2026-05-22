package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/middleware"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/utils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test_jwt_secret_for_middleware"

func init() {
	gin.SetMode(gin.TestMode)
}

type stubRevokedTokenRepository struct {
	revokedJTIs map[string]bool
	err         error
}

func (s *stubRevokedTokenRepository) Create(token *model.RevokedToken) error {
	return nil
}

func (s *stubRevokedTokenRepository) ExistsByJTI(jti string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.revokedJTIs[jti], nil
}

func newTestRouter(repo *stubRevokedTokenRepository, handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, repo))
	router.GET("/protected", handler)
	return router
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := newTestRouter(&stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "token_string"},
		{"Wrong prefix", "Basic token_string"},
		{"Missing token", "Bearer "},
		{"Empty Bearer", "Bearer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter(&stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tc.header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := newTestRouter(&stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token, err := utils.GenerateToken(123, "testuser", testJWTSecret, 24)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	var contextUserID uint64
	var contextUsername string
	var contextJTI string
	var contextExpiresAt time.Time

	router := newTestRouter(&stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}, func(c *gin.Context) {
		contextUserID = c.GetUint64("user_id")
		contextUsername = c.GetString("username")
		contextJTI = c.GetString("token_jti")
		expiresAtValue, ok := c.Get("token_expires_at")
		if !ok {
			t.Error("token_expires_at not found in context")
		}
		contextExpiresAt, ok = expiresAtValue.(time.Time)
		if !ok {
			t.Error("token_expires_at has unexpected type")
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if contextUserID != 123 {
		t.Errorf("Context user_id = %v, want 123", contextUserID)
	}
	if contextUsername != "testuser" {
		t.Errorf("Context username = %v, want testuser", contextUsername)
	}
	if contextJTI == "" {
		t.Error("Context token_jti should not be empty")
	}
	if contextExpiresAt.IsZero() {
		t.Error("Context token_expires_at should not be zero")
	}
}

func TestAuthMiddleware_RevokedToken(t *testing.T) {
	token, err := utils.GenerateToken(1, "testuser", testJWTSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	claims, err := utils.ParseToken(token, testJWTSecret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	router := newTestRouter(&stubRevokedTokenRepository{revokedJTIs: map[string]bool{claims.ID: true}}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_LegacyTokenWithoutJTI(t *testing.T) {
	legacyToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, utils.Claims{
		UserID:   1,
		Username: "legacy-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	router := newTestRouter(&stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+legacyToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	token, _ := utils.GenerateToken(1, "testuser", "different_secret_different_secret", 24)

	router := newTestRouter(&stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
