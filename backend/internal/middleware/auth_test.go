package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/middleware"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/utils"
	"strings"
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

// ========== 安全测试 ==========

func TestAuth_Security_MissingBearer(t *testing.T) {
	// 创建一个没有 Bearer 前缀的请求
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, &stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}))
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "invalid_token_format")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Security_EmptyToken(t *testing.T) {
	// 创建一个空 token 的请求
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, &stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}))
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Security_OnlyBearer(t *testing.T) {
	// 创建一个只有 Bearer 的请求
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, &stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}))
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Security_SQLInjection(t *testing.T) {
	// 创建一个包含 SQL 注入的 token
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, &stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}))
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer ' OR '1'='1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Security_XSSPayload(t *testing.T) {
	// 创建一个包含 XSS 的 token
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, &stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}))
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer <script>alert('xss')</script>")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Security_UnicodeToken(t *testing.T) {
	// 创建一个包含 Unicode 的 token
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, &stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}))
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer 测试token🎉")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Security_VeryLongHeader(t *testing.T) {
	// 创建一个超长的 Authorization header
	longToken := strings.Repeat("a", 10000)
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret, &stubRevokedTokenRepository{revokedJTIs: map[string]bool{}}))
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+longToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
