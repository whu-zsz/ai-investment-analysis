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
