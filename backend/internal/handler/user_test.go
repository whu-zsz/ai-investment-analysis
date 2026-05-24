package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/utils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type MockUserService struct {
	RegisterFunc      func(*request.RegisterRequest) (*model.User, error)
	LoginFunc         func(*request.LoginRequest) (*response.LoginResponse, error)
	LogoutFunc        func(uint64, string, time.Time) error
	GetProfileFunc    func(uint64) (*model.User, error)
	UpdateProfileFunc func(uint64, *request.UpdateProfileRequest) (*model.User, error)
}

func (m *MockUserService) Register(req *request.RegisterRequest) (*model.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(req)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserService) Login(req *request.LoginRequest) (*response.LoginResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(req)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserService) Logout(userID uint64, tokenJTI string, tokenExpiresAt time.Time) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(userID, tokenJTI, tokenExpiresAt)
	}
	return nil
}

func (m *MockUserService) GetProfile(userID uint64) (*model.User, error) {
	if m.GetProfileFunc != nil {
		return m.GetProfileFunc(userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserService) UpdateProfile(userID uint64, req *request.UpdateProfileRequest) (*model.User, error) {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(userID, req)
	}
	return nil, errors.New("not implemented")
}

const testJWTSecret = "test_jwt_secret_for_handler_test"

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRegister_Success(t *testing.T) {
	mockService := &MockUserService{
		RegisterFunc: func(req *request.RegisterRequest) (*model.User, error) {
			return &model.User{ID: 1, Username: req.Username, Email: req.Email}, nil
		},
	}

	h := handler.NewUserHandler(mockService)
	router := gin.New()
	router.POST("/register", h.Register)

	body := `{"username":"testuser","email":"test@example.com","password":"Test123456"}`
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", resp["code"])
	}
}

func TestRegister_InvalidRequest(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{"Empty body", ""},
		{"Missing password", `{"username":"test","email":"test@test.com"}`},
		{"Invalid email", `{"username":"test","email":"invalid","password":"123456"}`},
		{"Short password", `{"username":"test","email":"test@test.com","password":"123"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewUserHandler(&MockUserService{})
			router := gin.New()
			router.POST("/register", h.Register)

			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				t.Errorf("Expected error status, got %d", w.Code)
			}
		})
	}
}

func TestRegister_UsernameExists(t *testing.T) {
	h := handler.NewUserHandler(&MockUserService{RegisterFunc: func(req *request.RegisterRequest) (*model.User, error) {
		return nil, errors.New("username already exists")
	}})
	router := gin.New()
	router.POST("/register", h.Register)

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(`{"username":"existing","email":"test@test.com","password":"Test123456"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	h := handler.NewUserHandler(&MockUserService{LoginFunc: func(req *request.LoginRequest) (*response.LoginResponse, error) {
		token, _ := utils.GenerateToken(1, req.Username, testJWTSecret, 24)
		return &response.LoginResponse{Token: token, User: response.UserResponse{ID: 1, Username: req.Username, Email: "test@example.com"}}, nil
	}})
	router := gin.New()
	router.POST("/login", h.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"username":"testuser","password":"Test123456"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["token"] == "" {
		t.Error("Expected token in response")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h := handler.NewUserHandler(&MockUserService{LoginFunc: func(req *request.LoginRequest) (*response.LoginResponse, error) {
		return nil, errors.New("invalid username or password")
	}})
	router := gin.New()
	router.POST("/login", h.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"username":"testuser","password":"wrongpassword"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetProfile_Success(t *testing.T) {
	h := handler.NewUserHandler(&MockUserService{GetProfileFunc: func(userID uint64) (*model.User, error) {
		return &model.User{ID: userID, Username: "testuser", Email: "test@example.com", InvestmentPreference: "balanced", TotalProfit: decimal.NewFromInt(1000)}, nil
	}})
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, h.GetProfile)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/profile", nil))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetProfile_UserNotFound(t *testing.T) {
	h := handler.NewUserHandler(&MockUserService{GetProfileFunc: func(userID uint64) (*model.User, error) {
		return nil, errors.New("user not found")
	}})
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", uint64(999))
		c.Next()
	}, h.GetProfile)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/profile", nil))

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	h := handler.NewUserHandler(&MockUserService{UpdateProfileFunc: func(userID uint64, req *request.UpdateProfileRequest) (*model.User, error) {
		return &model.User{ID: userID, Username: "testuser", Phone: req.Phone, InvestmentPreference: *req.InvestmentPreference}, nil
	}})
	router := gin.New()
	router.PUT("/profile", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, h.UpdateProfile)

	req := httptest.NewRequest("PUT", "/profile", bytes.NewBufferString(`{"phone":"13800138000","investment_preference":"aggressive"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestLogout_Success(t *testing.T) {
	called := false
	var gotUserID uint64
	var gotJTI string
	var gotExpiresAt time.Time

	h := handler.NewUserHandler(&MockUserService{LogoutFunc: func(userID uint64, tokenJTI string, tokenExpiresAt time.Time) error {
		called = true
		gotUserID = userID
		gotJTI = tokenJTI
		gotExpiresAt = tokenExpiresAt
		return nil
	}})
	router := gin.New()
	router.POST("/logout", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Set("token_jti", "token-jti")
		c.Set("token_expires_at", time.Unix(1700000000, 0))
		c.Next()
	}, h.Logout)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/logout", nil))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !called {
		t.Fatal("Logout service was not called")
	}
	if gotUserID != 1 || gotJTI != "token-jti" || gotExpiresAt.IsZero() {
		t.Fatalf("unexpected logout args: userID=%d jti=%q expiresAt=%v", gotUserID, gotJTI, gotExpiresAt)
	}
}

func TestLogout_MissingContext(t *testing.T) {
	h := handler.NewUserHandler(&MockUserService{})
	router := gin.New()
	router.POST("/logout", h.Logout)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/logout", nil))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// ========== 安全测试 ==========

func TestUser_Security_SQLInjection_Register(t *testing.T) {
	sqlInjectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users; --",
		"' UNION SELECT * FROM users --",
		"admin'--",
	}

	for _, payload := range sqlInjectionPayloads {
		t.Run(payload, func(t *testing.T) {
			mockService := &MockUserService{
				RegisterFunc: func(req *request.RegisterRequest) (*model.User, error) {
					return &model.User{
						ID:       1,
						Username: req.Username,
						Email:    req.Email,
					}, nil
				},
			}

			h := handler.NewUserHandler(mockService)
			router := gin.New()
			router.POST("/register", h.Register)

			reqBody := fmt.Sprintf(`{
				"username": "%s",
				"email": "test@example.com",
				"password": "Test123456"
			}`, payload)

			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
				t.Errorf("Register with SQL injection '%s' returned %d", payload, w.Code)
			}
		})
	}
}

func TestUser_Security_XSS_Register(t *testing.T) {
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
		"<svg onload=alert('xss')>",
	}

	for _, payload := range xssPayloads {
		t.Run(payload, func(t *testing.T) {
			mockService := &MockUserService{
				RegisterFunc: func(req *request.RegisterRequest) (*model.User, error) {
					return &model.User{
						ID:       1,
						Username: req.Username,
						Email:    req.Email,
					}, nil
				},
			}

			h := handler.NewUserHandler(mockService)
			router := gin.New()
			router.POST("/register", h.Register)

			reqBody := fmt.Sprintf(`{
				"username": "%s",
				"email": "test@example.com",
				"password": "Test123456"
			}`, payload)

			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
				t.Errorf("Register with XSS '%s' returned %d", payload, w.Code)
			}
		})
	}
}

func TestUser_Security_SQLInjection_Login(t *testing.T) {
	sqlInjectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users; --",
		"admin'--",
	}

	for _, payload := range sqlInjectionPayloads {
		t.Run(payload, func(t *testing.T) {
			mockService := &MockUserService{
				LoginFunc: func(req *request.LoginRequest) (*response.LoginResponse, error) {
					return nil, errors.New("invalid username or password")
				},
			}

			h := handler.NewUserHandler(mockService)
			router := gin.New()
			router.POST("/login", h.Login)

			reqBody := fmt.Sprintf(`{
				"username": "%s",
				"password": "Test123456"
			}`, payload)

			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized && w.Code != http.StatusBadRequest {
				t.Errorf("Login with SQL injection '%s' returned %d", payload, w.Code)
			}
		})
	}
}

func TestUser_Security_SpecialChars_Profile(t *testing.T) {
	specialChars := []string{
		"test<script>alert('xss')</script>",
		"test'; DROP TABLE users; --",
		"test${7*7}",
		"test{{7*7}}",
		"test%00",
	}

	for _, chars := range specialChars {
		t.Run(chars, func(t *testing.T) {
			mockService := &MockUserService{
				UpdateProfileFunc: func(userID uint64, req *request.UpdateProfileRequest) (*model.User, error) {
					return &model.User{
						ID:                   userID,
						Username:             "testuser",
						Phone:                req.Phone,
						InvestmentPreference: *req.InvestmentPreference,
					}, nil
				},
			}

			h := handler.NewUserHandler(mockService)
			router := gin.New()
			router.PUT("/profile", func(c *gin.Context) {
				c.Set("user_id", uint64(1))
				c.Next()
			}, h.UpdateProfile)

			reqBody := fmt.Sprintf(`{
				"phone": "%s",
				"investment_preference": "balanced"
			}`, chars)

			req := httptest.NewRequest("PUT", "/profile", bytes.NewBufferString(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
				t.Errorf("Update profile with special chars '%s' returned %d", chars, w.Code)
			}
		})
	}
}

func TestUser_Security_OverflowID(t *testing.T) {
	overflowIDs := []string{
		"99999999999999999999",
		"18446744073709551615",
		"-1",
		"0",
	}

	for _, id := range overflowIDs {
		t.Run(id, func(t *testing.T) {
			mockService := &MockUserService{
				GetProfileFunc: func(userID uint64) (*model.User, error) {
					return nil, errors.New("invalid user id")
				},
			}

			h := handler.NewUserHandler(mockService)
			router := gin.New()
			router.GET("/transactions/:id", func(c *gin.Context) {
				c.Set("user_id", uint64(1))
				c.Next()
			}, h.GetProfile)

			req := httptest.NewRequest("GET", "/transactions/"+id, nil)
			req.Header.Set("Authorization", "Bearer valid_token")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized && w.Code != http.StatusNotFound {
				t.Errorf("Request with overflow ID '%s' returned %d", id, w.Code)
			}
		})
	}
}

// ========== 性能测试 ==========

func BenchmarkUser_Register(b *testing.B) {
	// 创建 mock 服务
	mockService := &MockUserService{
		RegisterFunc: func(req *request.RegisterRequest) (*model.User, error) {
			return &model.User{
				ID:       1,
				Username: req.Username,
				Email:    req.Email,
			}, nil
		},
	}

	// 创建路由
	router := gin.New()
	h := handler.NewUserHandler(mockService)
	router.POST("/api/v1/auth/register", h.Register)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reqBody := fmt.Sprintf(`{
			"username": "user_%d",
			"email": "user_%d@example.com",
			"password": "Test123456"
		}`, i, i)

		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Register returned %d", w.Code)
		}
	}
}

func BenchmarkUser_Login(b *testing.B) {
	// 创建 mock 服务
	mockService := &MockUserService{
		LoginFunc: func(req *request.LoginRequest) (*response.LoginResponse, error) {
			return &response.LoginResponse{
				Token: "test_token",
				User: response.UserResponse{
					ID:       1,
					Username: req.Username,
				},
			}, nil
		},
	}

	// 创建路由
	router := gin.New()
	h := handler.NewUserHandler(mockService)
	router.POST("/api/v1/auth/login", h.Login)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reqBody := `{
			"username": "testuser",
			"password": "Test123456"
		}`

		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Login returned %d", w.Code)
		}
	}
}

func BenchmarkUser_GetProfile(b *testing.B) {
	// 创建 mock 服务
	mockService := &MockUserService{
		GetProfileFunc: func(userID uint64) (*model.User, error) {
			return &model.User{
				ID:       userID,
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
	}

	// 创建路由
	router := gin.New()
	h := handler.NewUserHandler(mockService)
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		h.GetProfile(c)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("GetProfile returned %d", w.Code)
		}
	}
}
