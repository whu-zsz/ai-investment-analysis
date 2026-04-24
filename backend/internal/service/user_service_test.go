package service_test

import (
	"errors"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockUserRepository 实现 UserRepository 接口用于测试
type MockUserRepository struct {
	users         map[uint64]*model.User
	nextID        uint64
	errOnCreate   error
	errOnFindByID error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[uint64]*model.User),
		nextID: 1,
	}
}

func (m *MockUserRepository) Create(user *model.User) error {
	if m.errOnCreate != nil {
		return m.errOnCreate
	}
	user.ID = m.nextID
	// 模拟 GORM 的 default:true 行为
	// 如果 IsActive 没有被显式设置（false），则设为 true
	// 注意：这里我们假设新注册的用户默认是活跃的
	m.users[m.nextID] = user
	m.nextID++
	return nil
}

// SetUserActive 设置用户活跃状态（用于测试）
func (m *MockUserRepository) SetUserActive(id uint64, active bool) {
	if user, ok := m.users[id]; ok {
		user.IsActive = active
	}
}

func (m *MockUserRepository) FindByID(id uint64) (*model.User, error) {
	if m.errOnFindByID != nil {
		return nil, m.errOnFindByID
	}
	user, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByUsername(username string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) Update(user *model.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(id uint64) error {
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(id uint64) error {
	return nil
}

func (m *MockUserRepository) UpdateTotalProfit(id uint64, profit decimal.Decimal) error {
	return nil
}

// 测试配置
func getTestJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:      "test_secret_key",
		ExpireHours: 24,
	}
}

// TestUserService_Register_Success 测试用户注册成功
func TestUserService_Register_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	req := &request.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123",
	}

	user, err := svc.Register(req)

	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if user.ID == 0 {
		t.Error("Register() user ID should not be 0")
	}

	if user.Username != req.Username {
		t.Errorf("Username = %v, want %v", user.Username, req.Username)
	}

	if user.Email != req.Email {
		t.Errorf("Email = %v, want %v", user.Email, req.Email)
	}

	if user.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}

	if user.PasswordHash == req.Password {
		t.Error("PasswordHash should not equal plain password")
	}
}

// TestUserService_Register_UsernameExists 测试用户名已存在
func TestUserService_Register_UsernameExists(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	// 先注册一个用户
	repo.Create(&model.User{Username: "existing", Email: "existing@example.com", PasswordHash: "hash"})

	req := &request.RegisterRequest{
		Username: "existing",
		Email:    "new@example.com",
		Password: "Password123",
	}

	_, err := svc.Register(req)

	if err == nil {
		t.Error("Register() should return error for existing username")
	}

	if err.Error() != "username already exists" {
		t.Errorf("Error message = %v, want 'username already exists'", err.Error())
	}
}

// TestUserService_Register_EmailExists 测试邮箱已存在
func TestUserService_Register_EmailExists(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	// 先注册一个用户
	repo.Create(&model.User{Username: "user1", Email: "existing@example.com", PasswordHash: "hash"})

	req := &request.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "Password123",
	}

	_, err := svc.Register(req)

	if err == nil {
		t.Error("Register() should return error for existing email")
	}

	if err.Error() != "email already exists" {
		t.Errorf("Error message = %v, want 'email already exists'", err.Error())
	}
}

// TestUserService_Login_Success 测试登录成功
func TestUserService_Login_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	// 注册用户
	req := &request.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123",
	}
	user, _ := svc.Register(req)
	// 设置用户为活跃状态（模拟数据库默认值）
	repo.SetUserActive(user.ID, true)

	// 登录
	loginReq := &request.LoginRequest{
		Username: "testuser",
		Password: "Password123",
	}

	resp, err := svc.Login(loginReq)

	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if resp.Token == "" {
		t.Error("Login() token should not be empty")
	}

	if resp.User.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", resp.User.Username)
	}
}

// TestUserService_Login_WrongPassword 测试密码错误
func TestUserService_Login_WrongPassword(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	// 注册用户
	user, _ := svc.Register(&request.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "CorrectPassword",
	})
	repo.SetUserActive(user.ID, true)

	// 使用错误密码登录
	loginReq := &request.LoginRequest{
		Username: "testuser",
		Password: "WrongPassword",
	}

	_, err := svc.Login(loginReq)

	if err == nil {
		t.Error("Login() should return error for wrong password")
	}

	if err.Error() != "invalid username or password" {
		t.Errorf("Error message = %v, want 'invalid username or password'", err.Error())
	}
}

// TestUserService_Login_UserNotFound 测试用户不存在
func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	loginReq := &request.LoginRequest{
		Username: "nonexistent",
		Password: "Password123",
	}

	_, err := svc.Login(loginReq)

	if err == nil {
		t.Error("Login() should return error for non-existent user")
	}

	if err.Error() != "invalid username or password" {
		t.Errorf("Error message = %v, want 'invalid username or password'", err.Error())
	}
}

// TestUserService_GetProfile_Success 测试获取用户信息成功
func TestUserService_GetProfile_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	// 创建用户
	user := &model.User{
		Username:             "testuser",
		Email:                "test@example.com",
		InvestmentPreference: "aggressive",
	}
	repo.Create(user)

	// 获取用户信息
	result, err := svc.GetProfile(user.ID)

	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", result.Username)
	}
}

// TestUserService_GetProfile_UserNotFound 测试用户不存在
func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	_, err := svc.GetProfile(999)

	if err == nil {
		t.Error("GetProfile() should return error for non-existent user")
	}
}

// TestUserService_UpdateProfile_Success 测试更新用户信息成功
func TestUserService_UpdateProfile_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	// 创建用户
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	repo.Create(user)

	// 更新用户信息
	phone := "13800138000"
	preference := "aggressive"
	req := &request.UpdateProfileRequest{
		Phone:                &phone,
		InvestmentPreference: &preference,
	}

	result, err := svc.UpdateProfile(user.ID, req)

	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}

	if result.Phone == nil || *result.Phone != phone {
		t.Errorf("Phone = %v, want %v", result.Phone, phone)
	}

	if result.InvestmentPreference != preference {
		t.Errorf("InvestmentPreference = %v, want %v", result.InvestmentPreference, preference)
	}
}

// TestUserService_UpdateProfile_UserNotFound 测试更新不存在的用户
func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	preference := "aggressive"
	req := &request.UpdateProfileRequest{
		InvestmentPreference: &preference,
	}

	_, err := svc.UpdateProfile(999, req)

	if err == nil {
		t.Error("UpdateProfile() should return error for non-existent user")
	}
}

// TestUserService_Login_InactiveUser 测试用户账户已停用
func TestUserService_Login_InactiveUser(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo, getTestJWTConfig())

	// 先注册一个正常用户获取正确的密码哈希
	svc.Register(&request.RegisterRequest{
		Username: "temp",
		Email:    "temp@example.com",
		Password: "Password123",
	})

	tempUser, _ := repo.FindByUsername("temp")

	// 创建停用的用户（使用相同的密码哈希）
	inactiveUser := &model.User{
		Username:     "inactive",
		Email:        "inactive@example.com",
		PasswordHash: tempUser.PasswordHash,
		IsActive:     false,
	}
	repo.Create(inactiveUser)

	loginReq := &request.LoginRequest{
		Username: "inactive",
		Password: "Password123",
	}

	_, err := svc.Login(loginReq)

	if err == nil {
		t.Error("Login() should return error for inactive user")
	}

	if err.Error() != "user account is deactivated" {
		t.Errorf("Error message = %v, want 'user account is deactivated'", err.Error())
	}
}

// TestUserService_Register_DatabaseError 测试数据库错误
func TestUserService_Register_DatabaseError(t *testing.T) {
	repo := NewMockUserRepository()
	repo.errOnCreate = errors.New("database connection error")
	svc := service.NewUserService(repo, getTestJWTConfig())

	req := &request.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123",
	}

	// 首先注册一次使其用户名存在
	_, _ = svc.Register(req)

	// 由于我们设置了错误，所以应该返回错误
	_, err := svc.Register(req)

	if err == nil {
		t.Error("Register() should return error on database error")
	}
}
