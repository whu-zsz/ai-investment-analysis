package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockDB 模拟 GORM DB 对象的行为
// 由于 GORM 的复杂性，我们创建一个简单的内存存储模拟

// InMemoryUserRepository 内存用户仓储用于测试
type InMemoryUserRepository struct {
	users    map[uint64]*model.User
	nextID   uint64
	username map[string]*model.User
	email    map[string]*model.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:    make(map[uint64]*model.User),
		nextID:   1,
		username: make(map[string]*model.User),
		email:    make(map[string]*model.User),
	}
}

func (r *InMemoryUserRepository) Create(user *model.User) error {
	user.ID = r.nextID
	r.users[r.nextID] = user
	r.username[user.Username] = user
	r.email[user.Email] = user
	r.nextID++
	return nil
}

func (r *InMemoryUserRepository) FindByID(id uint64) (*model.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) FindByUsername(username string) (*model.User, error) {
	user, ok := r.username[username]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (*model.User, error) {
	user, ok := r.email[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) Update(user *model.User) error {
	r.users[user.ID] = user
	r.username[user.Username] = user
	r.email[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) Delete(id uint64) error {
	user, ok := r.users[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.users, id)
	delete(r.username, user.Username)
	delete(r.email, user.Email)
	return nil
}

func (r *InMemoryUserRepository) UpdateLastLogin(id uint64) error {
	_, ok := r.users[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *InMemoryUserRepository) UpdateTotalProfit(id uint64, profit decimal.Decimal) error {
	user, ok := r.users[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	user.TotalProfit = profit
	return nil
}

// TestUserRepository_Create 测试创建用户
func TestUserRepository_Create(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if user.ID == 0 {
		t.Error("Create() should assign ID")
	}

	// 验证可以通过 ID 查找
	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", found.Username)
	}
}

// TestUserRepository_FindByUsername 测试通过用户名查找
func TestUserRepository_FindByUsername(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// 创建用户
	repo.Create(&model.User{Username: "user1", Email: "user1@test.com", PasswordHash: "hash"})
	repo.Create(&model.User{Username: "user2", Email: "user2@test.com", PasswordHash: "hash"})

	// 查找存在的用户
	user, err := repo.FindByUsername("user1")
	if err != nil {
		t.Fatalf("FindByUsername() error = %v", err)
	}
	if user.Username != "user1" {
		t.Errorf("Username = %v, want user1", user.Username)
	}

	// 查找不存在的用户
	_, err = repo.FindByUsername("nonexistent")
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByUsername() should return ErrRecordNotFound for non-existent user")
	}
}

// TestUserRepository_FindByEmail 测试通过邮箱查找
func TestUserRepository_FindByEmail(t *testing.T) {
	repo := NewInMemoryUserRepository()

	repo.Create(&model.User{Username: "user1", Email: "user1@test.com", PasswordHash: "hash"})

	// 查找存在的邮箱
	user, err := repo.FindByEmail("user1@test.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if user.Email != "user1@test.com" {
		t.Errorf("Email = %v, want user1@test.com", user.Email)
	}

	// 查找不存在的邮箱
	_, err = repo.FindByEmail("nonexistent@test.com")
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByEmail() should return ErrRecordNotFound for non-existent email")
	}
}

// TestUserRepository_Update 测试更新用户
func TestUserRepository_Update(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &model.User{
		Username:             "testuser",
		Email:                "test@example.com",
		PasswordHash:         "hash",
		InvestmentPreference: "balanced",
	}
	repo.Create(user)

	// 更新用户
	user.InvestmentPreference = "aggressive"
	err := repo.Update(user)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// 验证更新
	found, _ := repo.FindByID(user.ID)
	if found.InvestmentPreference != "aggressive" {
		t.Errorf("InvestmentPreference = %v, want aggressive", found.InvestmentPreference)
	}
}

// TestUserRepository_Delete 测试删除用户
func TestUserRepository_Delete(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &model.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"}
	repo.Create(user)

	// 删除用户
	err := repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证删除
	_, err = repo.FindByID(user.ID)
	if err != gorm.ErrRecordNotFound {
		t.Error("Delete() should remove user")
	}

	// 删除不存在的用户
	err = repo.Delete(999)
	if err != gorm.ErrRecordNotFound {
		t.Error("Delete() should return error for non-existent user")
	}
}

// TestUserRepository_UpdateTotalProfit 测试更新总盈亏
func TestUserRepository_UpdateTotalProfit(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		TotalProfit:  decimal.Zero,
	}
	repo.Create(user)

	// 更新总盈亏
	profit := decimal.NewFromFloat(5000.00)
	err := repo.UpdateTotalProfit(user.ID, profit)
	if err != nil {
		t.Fatalf("UpdateTotalProfit() error = %v", err)
	}

	// 验证更新
	found, _ := repo.FindByID(user.ID)
	if !found.TotalProfit.Equals(profit) {
		t.Errorf("TotalProfit = %v, want %v", found.TotalProfit, profit)
	}
}

// TestUserRepository_FindByID_NotFound 测试查找不存在的用户
func TestUserRepository_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	_, err := repo.FindByID(999)
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByID() should return ErrRecordNotFound for non-existent ID")
	}
}

// TestUserRepository_MultipleUsers 测试多用户场景
func TestUserRepository_MultipleUsers(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// 创建多个用户
	users := []*model.User{
		{Username: "user1", Email: "user1@test.com", PasswordHash: "hash1"},
		{Username: "user2", Email: "user2@test.com", PasswordHash: "hash2"},
		{Username: "user3", Email: "user3@test.com", PasswordHash: "hash3"},
	}

	for _, u := range users {
		repo.Create(u)
	}

	// 验证所有用户都可以找到
	for _, u := range users {
		found, err := repo.FindByID(u.ID)
		if err != nil {
			t.Errorf("FindByID(%d) error = %v", u.ID, err)
		}
		if found.Username != u.Username {
			t.Errorf("Username = %v, want %v", found.Username, u.Username)
		}
	}
}

// 确保 InMemoryUserRepository 实现了 UserRepository 接口
var _ repository.UserRepository = (*InMemoryUserRepository)(nil)

// TestUserRepository_Interface 测试接口实现
func TestUserRepository_Interface(t *testing.T) {
	// 这个测试确保我们的 mock 实现了所有必要的方法
	var repo repository.UserRepository = NewInMemoryUserRepository()

	user := &model.User{Username: "test", Email: "test@test.com", PasswordHash: "hash"}

	// 测试所有接口方法
	_ = repo.Create(user)
	_, _ = repo.FindByID(user.ID)
	_, _ = repo.FindByUsername("test")
	_, _ = repo.FindByEmail("test@test.com")
	_ = repo.Update(user)
	_ = repo.UpdateLastLogin(user.ID)
	_ = repo.UpdateTotalProfit(user.ID, decimal.Zero)
	_ = repo.Delete(user.ID)
}
