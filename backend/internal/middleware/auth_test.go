package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/middleware"
	"stock-analysis-backend/internal/utils"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const testJWTSecret = "test_jwt_secret_for_middleware"

func init() {
	gin.SetMode(gin.TestMode)
}

// TestAuthMiddleware_MissingHeader 测试缺少 Authorization Header
func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if w.Body.String() == "" {
		t.Error("Response body should not be empty")
	}
}

// TestAuthMiddleware_InvalidFormat 测试无效的 Authorization 格式
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
			router := gin.New()
			router.Use(middleware.AuthMiddleware(testJWTSecret))
			router.GET("/protected", func(c *gin.Context) {
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

// TestAuthMiddleware_InvalidToken 测试无效 Token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
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

// TestAuthMiddleware_ValidToken 测试有效 Token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	// 生成有效 token
	userID := uint64(123)
	username := "testuser"
	token, err := utils.GenerateToken(userID, username, testJWTSecret, 24)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		// 验证 context 中存储的用户信息
		ctxUserID, exists := c.Get("user_id")
		if !exists {
			t.Error("user_id not found in context")
		}
		ctxUsername, _ := c.Get("username")

		c.JSON(http.StatusOK, gin.H{
			"user_id":  ctxUserID,
			"username": ctxUsername,
		})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestAuthMiddleware_WrongSecret 测试使用不同密钥签名的 Token
func TestAuthMiddleware_WrongSecret(t *testing.T) {
	// 使用不同密钥生成 token
	token, _ := utils.GenerateToken(1, "testuser", "different_secret", 24)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
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

// TestAuthMiddleware_ContextValues 测试 Context 中存储的用户信息
func TestAuthMiddleware_ContextValues(t *testing.T) {
	testCases := []struct {
		name     string
		userID   uint64
		username string
	}{
		{"User 1", 1, "alice"},
		{"User 2", 999, "bob"},
		{"Chinese user", 100, "测试用户"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, _ := utils.GenerateToken(tc.userID, tc.username, testJWTSecret, 24)

			var contextUserID uint64
			var contextUsername string

			router := gin.New()
			router.Use(middleware.AuthMiddleware(testJWTSecret))
			router.GET("/protected", func(c *gin.Context) {
				contextUserID = c.GetUint64("user_id")
				contextUsername = c.GetString("username")
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if contextUserID != tc.userID {
				t.Errorf("Context user_id = %v, want %v", contextUserID, tc.userID)
			}
			if contextUsername != tc.username {
				t.Errorf("Context username = %v, want %v", contextUsername, tc.username)
			}
		})
	}
}

// ========== 安全测试 ==========

func TestAuth_Security_MissingBearer(t *testing.T) {
	// 创建一个没有 Bearer 前缀的请求
	router := gin.New()
	router.Use(middleware.AuthMiddleware(testJWTSecret))
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
	router.Use(middleware.AuthMiddleware(testJWTSecret))
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
	router.Use(middleware.AuthMiddleware(testJWTSecret))
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
	router.Use(middleware.AuthMiddleware(testJWTSecret))
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
	router.Use(middleware.AuthMiddleware(testJWTSecret))
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
	router.Use(middleware.AuthMiddleware(testJWTSecret))
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
	router.Use(middleware.AuthMiddleware(testJWTSecret))
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
