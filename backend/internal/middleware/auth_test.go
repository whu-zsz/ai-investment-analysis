package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/middleware"
	"stock-analysis-backend/internal/utils"
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
