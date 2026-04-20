package utils_test

import (
	"stock-analysis-backend/internal/utils"
	"testing"
)

// TestHashPassword_Success 测试密码哈希成功
func TestHashPassword_Success(t *testing.T) {
	password := "TestPassword123"
	hash, err := utils.HashPassword(password)

	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty string")
	}

	if hash == password {
		t.Error("HashPassword() returned plain password, should be hashed")
	}
}

// TestHashPassword_DifferentPasswords 测试不同密码生成不同哈希
func TestHashPassword_DifferentPasswords(t *testing.T) {
	password1 := "Password1"
	password2 := "Password2"

	hash1, _ := utils.HashPassword(password1)
	hash2, _ := utils.HashPassword(password2)

	if hash1 == hash2 {
		t.Error("Different passwords should produce different hashes")
	}
}

// TestHashPassword_SamePasswordDifferentHash 测试相同密码每次生成不同哈希（bcrypt 特性）
func TestHashPassword_SamePasswordDifferentHash(t *testing.T) {
	password := "SamePassword123"

	hash1, _ := utils.HashPassword(password)
	hash2, _ := utils.HashPassword(password)

	// bcrypt 每次生成的哈希不同（因为有随机盐）
	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to salt")
	}
}

// TestCheckPassword_CorrectPassword 测试正确密码验证
func TestCheckPassword_CorrectPassword(t *testing.T) {
	password := "CorrectPassword123"
	hash, _ := utils.HashPassword(password)

	if !utils.CheckPassword(password, hash) {
		t.Error("CheckPassword() should return true for correct password")
	}
}

// TestCheckPassword_WrongPassword 测试错误密码验证
func TestCheckPassword_WrongPassword(t *testing.T) {
	password := "CorrectPassword123"
	wrongPassword := "WrongPassword456"
	hash, _ := utils.HashPassword(password)

	if utils.CheckPassword(wrongPassword, hash) {
		t.Error("CheckPassword() should return false for wrong password")
	}
}

// TestCheckPassword_EmptyPassword 测试空密码
func TestCheckPassword_EmptyPassword(t *testing.T) {
	password := "NonEmptyPassword"
	hash, _ := utils.HashPassword(password)

	if utils.CheckPassword("", hash) {
		t.Error("CheckPassword() should return false for empty password")
	}
}

// TestCheckPassword_InvalidHash 测试无效哈希
func TestCheckPassword_InvalidHash(t *testing.T) {
	password := "TestPassword"
	invalidHash := "invalid_hash_string"

	if utils.CheckPassword(password, invalidHash) {
		t.Error("CheckPassword() should return false for invalid hash")
	}
}

// TestHashPassword_EmptyPassword 测试空密码哈希
func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := utils.HashPassword("")

	if err != nil {
		t.Fatalf("HashPassword() with empty password should not error, got %v", err)
	}

	if !utils.CheckPassword("", hash) {
		t.Error("Empty password should match its hash")
	}
}
