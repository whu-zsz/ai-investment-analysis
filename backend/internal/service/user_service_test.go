package service_test

import (
	"errors"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	users         map[uint64]*model.User
	nextID        uint64
	errOnCreate   error
	errOnFindByID error
}

type MockRevokedTokenRepository struct {
	tokens        map[string]*model.RevokedToken
	createErr     error
	duplicateJTIs map[string]bool
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[uint64]*model.User),
		nextID: 1,
	}
}

func NewMockRevokedTokenRepository() *MockRevokedTokenRepository {
	return &MockRevokedTokenRepository{
		tokens:        make(map[string]*model.RevokedToken),
		duplicateJTIs: make(map[string]bool),
	}
}

func (m *MockUserRepository) Create(user *model.User) error {
	if m.errOnCreate != nil {
		return m.errOnCreate
	}
	user.ID = m.nextID
	m.users[m.nextID] = user
	m.nextID++
	return nil
}

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

func (m *MockRevokedTokenRepository) Create(token *model.RevokedToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.duplicateJTIs[token.JTI] || m.tokens[token.JTI] != nil {
		return gorm.ErrDuplicatedKey
	}
	m.tokens[token.JTI] = token
	return nil
}

func (m *MockRevokedTokenRepository) ExistsByJTI(jti string) (bool, error) {
	_, ok := m.tokens[jti]
	return ok, nil
}

func getTestJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:      "test_secret_key_with_32_characters_minimum",
		ExpireHours: 24,
	}
}

func newTestUserService(repo *MockUserRepository, revokedRepo *MockRevokedTokenRepository) service.UserService {
	return service.NewUserService(repo, revokedRepo, getTestJWTConfig())
}

func TestUserService_Register_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())

	user, err := svc.Register(&request.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123",
	})

	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.ID == 0 {
		t.Error("Register() user ID should not be 0")
	}
	if user.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", user.Email)
	}
	if user.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}
}

func TestUserService_Register_UsernameExists(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())
	repo.Create(&model.User{Username: "existing", Email: "existing@example.com", PasswordHash: "hash"})

	_, err := svc.Register(&request.RegisterRequest{Username: "existing", Email: "new@example.com", Password: "Password123"})
	if err == nil {
		t.Error("Register() should return error for existing username")
	}
	if err.Error() != "username already exists" {
		t.Errorf("Error message = %v, want 'username already exists'", err.Error())
	}
}

func TestUserService_Register_EmailExists(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())
	repo.Create(&model.User{Username: "user1", Email: "existing@example.com", PasswordHash: "hash"})

	_, err := svc.Register(&request.RegisterRequest{Username: "newuser", Email: "existing@example.com", Password: "Password123"})
	if err == nil {
		t.Error("Register() should return error for existing email")
	}
	if err.Error() != "email already exists" {
		t.Errorf("Error message = %v, want 'email already exists'", err.Error())
	}
}

func TestUserService_Login_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())

	user, _ := svc.Register(&request.RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "Password123"})
	repo.SetUserActive(user.ID, true)

	resp, err := svc.Login(&request.LoginRequest{Username: "testuser", Password: "Password123"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.Token == "" {
		t.Error("Login() token should not be empty")
	}
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())

	user, _ := svc.Register(&request.RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "CorrectPassword"})
	repo.SetUserActive(user.ID, true)

	_, err := svc.Login(&request.LoginRequest{Username: "testuser", Password: "WrongPassword"})
	if err == nil {
		t.Error("Login() should return error for wrong password")
	}
	if err.Error() != "invalid username or password" {
		t.Errorf("Error message = %v, want 'invalid username or password'", err.Error())
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())

	_, err := svc.Login(&request.LoginRequest{Username: "nonexistent", Password: "Password123"})
	if err == nil {
		t.Error("Login() should return error for non-existent user")
	}
}

func TestUserService_GetProfile_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())
	repo.Create(&model.User{Username: "testuser", Email: "test@example.com", InvestmentPreference: "aggressive"})

	result, err := svc.GetProfile(1)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if result.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", result.Username)
	}
}

func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	svc := newTestUserService(NewMockUserRepository(), NewMockRevokedTokenRepository())
	if _, err := svc.GetProfile(999); err == nil {
		t.Error("GetProfile() should return error for non-existent user")
	}
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())
	repo.Create(&model.User{Username: "testuser", Email: "test@example.com"})

	phone := "13800138000"
	preference := "aggressive"
	result, err := svc.UpdateProfile(1, &request.UpdateProfileRequest{Phone: &phone, InvestmentPreference: &preference})
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

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	svc := newTestUserService(NewMockUserRepository(), NewMockRevokedTokenRepository())
	preference := "aggressive"
	if _, err := svc.UpdateProfile(999, &request.UpdateProfileRequest{InvestmentPreference: &preference}); err == nil {
		t.Error("UpdateProfile() should return error for non-existent user")
	}
}

func TestUserService_Login_InactiveUser(t *testing.T) {
	repo := NewMockUserRepository()
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())

	svc.Register(&request.RegisterRequest{Username: "temp", Email: "temp@example.com", Password: "Password123"})
	tempUser, _ := repo.FindByUsername("temp")
	repo.Create(&model.User{Username: "inactive", Email: "inactive@example.com", PasswordHash: tempUser.PasswordHash, IsActive: false})

	_, err := svc.Login(&request.LoginRequest{Username: "inactive", Password: "Password123"})
	if err == nil {
		t.Error("Login() should return error for inactive user")
	}
	if err.Error() != "user account is deactivated" {
		t.Errorf("Error message = %v, want 'user account is deactivated'", err.Error())
	}
}

func TestUserService_Register_DatabaseError(t *testing.T) {
	repo := NewMockUserRepository()
	repo.errOnCreate = errors.New("database connection error")
	svc := newTestUserService(repo, NewMockRevokedTokenRepository())

	_, err := svc.Register(&request.RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "Password123"})
	if err == nil {
		t.Error("Register() should return error on database error")
	}
}

func TestUserService_Logout_Success(t *testing.T) {
	revokedRepo := NewMockRevokedTokenRepository()
	svc := newTestUserService(NewMockUserRepository(), revokedRepo)
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := svc.Logout(1, "token-jti", expiresAt); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	token, ok := revokedRepo.tokens["token-jti"]
	if !ok {
		t.Fatal("revoked token not written")
	}
	if token.UserID != 1 {
		t.Errorf("revoked token user_id = %d, want 1", token.UserID)
	}
}

func TestUserService_Logout_RequiresJTI(t *testing.T) {
	svc := newTestUserService(NewMockUserRepository(), NewMockRevokedTokenRepository())
	if err := svc.Logout(1, "   ", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("Logout() error = nil, want error")
	}
}

func TestUserService_Logout_IdempotentOnDuplicate(t *testing.T) {
	revokedRepo := NewMockRevokedTokenRepository()
	svc := newTestUserService(NewMockUserRepository(), revokedRepo)
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := svc.Logout(1, "duplicate-jti", expiresAt); err != nil {
		t.Fatalf("first Logout() error = %v", err)
	}
	revokedRepo.duplicateJTIs["duplicate-jti"] = true
	if err := svc.Logout(1, "duplicate-jti", expiresAt); err != nil {
		t.Fatalf("second Logout() error = %v, want nil", err)
	}
}
