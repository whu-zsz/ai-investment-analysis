package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// MockUserService 实现 UserService 接口用于测试
type MockUserService struct {
	RegisterFunc      func(*request.RegisterRequest) (*model.User, error)
	LoginFunc         func(*request.LoginRequest) (*response.LoginResponse, error)
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

// TestRegister_Success 测试注册成功
func TestRegister_Success(t *testing.T) {
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

// TestRegister_InvalidRequest 测试无效注册请求
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
			mockService := &MockUserService{}
			h := handler.NewUserHandler(mockService)
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

// TestRegister_UsernameExists 测试用户名已存在
func TestRegister_UsernameExists(t *testing.T) {
	mockService := &MockUserService{
		RegisterFunc: func(req *request.RegisterRequest) (*model.User, error) {
			return nil, errors.New("username already exists")
		},
	}

	h := handler.NewUserHandler(mockService)
	router := gin.New()
	router.POST("/register", h.Register)

	body := `{"username":"existing","email":"test@test.com","password":"Test123456"}`
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestLogin_Success 测试登录成功
func TestLogin_Success(t *testing.T) {
	mockService := &MockUserService{
		LoginFunc: func(req *request.LoginRequest) (*response.LoginResponse, error) {
			token, _ := utils.GenerateToken(1, req.Username, testJWTSecret, 24)
			return &response.LoginResponse{
				Token: token,
				User: response.UserResponse{
					ID:       1,
					Username: req.Username,
					Email:    "test@example.com",
				},
			}, nil
		},
	}

	h := handler.NewUserHandler(mockService)
	router := gin.New()
	router.POST("/login", h.Login)

	body := `{"username":"testuser","password":"Test123456"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
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

// TestLogin_InvalidCredentials 测试登录凭证错误
func TestLogin_InvalidCredentials(t *testing.T) {
	mockService := &MockUserService{
		LoginFunc: func(req *request.LoginRequest) (*response.LoginResponse, error) {
			return nil, errors.New("invalid username or password")
		},
	}

	h := handler.NewUserHandler(mockService)
	router := gin.New()
	router.POST("/login", h.Login)

	body := `{"username":"testuser","password":"wrongpassword"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestGetProfile_Success 测试获取用户信息成功
func TestGetProfile_Success(t *testing.T) {
	mockService := &MockUserService{
		GetProfileFunc: func(userID uint64) (*model.User, error) {
			return &model.User{
				ID:                   userID,
				Username:             "testuser",
				Email:                "test@example.com",
				InvestmentPreference: "balanced",
				TotalProfit:          decimal.NewFromInt(1000),
			}, nil
		},
	}

	h := handler.NewUserHandler(mockService)
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, h.GetProfile)

	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestGetProfile_UserNotFound 测试用户不存在
func TestGetProfile_UserNotFound(t *testing.T) {
	mockService := &MockUserService{
		GetProfileFunc: func(userID uint64) (*model.User, error) {
			return nil, errors.New("user not found")
		},
	}

	h := handler.NewUserHandler(mockService)
	router := gin.New()
	router.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", uint64(999))
		c.Next()
	}, h.GetProfile)

	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestUpdateProfile_Success 测试更新用户信息成功
func TestUpdateProfile_Success(t *testing.T) {
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

	body := `{"phone":"13800138000","investment_preference":"aggressive"}`
	req := httptest.NewRequest("PUT", "/profile", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestLogout_Success 测试登出成功
func TestLogout_Success(t *testing.T) {
	mockService := &MockUserService{}
	h := handler.NewUserHandler(mockService)
	router := gin.New()
	router.POST("/logout", h.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
