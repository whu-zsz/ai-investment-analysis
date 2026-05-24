package repository_test

import (
	"errors"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"gorm.io/gorm"
)

type inMemoryRevokedTokenRepository struct {
	tokens map[string]*model.RevokedToken
}

func newInMemoryRevokedTokenRepository() *inMemoryRevokedTokenRepository {
	return &inMemoryRevokedTokenRepository{tokens: make(map[string]*model.RevokedToken)}
}

func (r *inMemoryRevokedTokenRepository) Create(token *model.RevokedToken) error {
	if _, exists := r.tokens[token.JTI]; exists {
		return gorm.ErrDuplicatedKey
	}
	r.tokens[token.JTI] = token
	return nil
}

func (r *inMemoryRevokedTokenRepository) ExistsByJTI(jti string) (bool, error) {
	_, exists := r.tokens[jti]
	return exists, nil
}

func TestInMemoryRevokedTokenRepositoryCreateAndExistsByJTI(t *testing.T) {
	repo := newInMemoryRevokedTokenRepository()
	record := &model.RevokedToken{
		UserID:         1,
		JTI:            "token-jti-1",
		TokenExpiresAt: time.Now().Add(24 * time.Hour),
		RevokedAt:      time.Now(),
		Reason:         "logout",
	}

	if err := repo.Create(record); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	exists, err := repo.ExistsByJTI("token-jti-1")
	if err != nil {
		t.Fatalf("ExistsByJTI() error = %v", err)
	}
	if !exists {
		t.Fatal("ExistsByJTI() = false, want true")
	}
}

func TestInMemoryRevokedTokenRepositoryExistsByJTIMissing(t *testing.T) {
	repo := newInMemoryRevokedTokenRepository()

	exists, err := repo.ExistsByJTI("missing-jti")
	if err != nil {
		t.Fatalf("ExistsByJTI() error = %v", err)
	}
	if exists {
		t.Fatal("ExistsByJTI() = true, want false")
	}
}

func TestInMemoryRevokedTokenRepositoryRejectsDuplicateJTI(t *testing.T) {
	repo := newInMemoryRevokedTokenRepository()
	record := &model.RevokedToken{
		UserID:         1,
		JTI:            "duplicate-jti",
		TokenExpiresAt: time.Now().Add(24 * time.Hour),
		RevokedAt:      time.Now(),
		Reason:         "logout",
	}

	if err := repo.Create(record); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if err := repo.Create(record); err == nil {
		t.Fatal("second Create() error = nil, want duplicate error")
	} else if !repository.IsDuplicateRevokedTokenError(err) {
		t.Fatalf("duplicate error classification failed: %v", err)
	}
}

func TestIsDuplicateRevokedTokenError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "gorm duplicated key", err: gorm.ErrDuplicatedKey, want: true},
		{name: "generic duplicate message", err: errors.New("duplicate entry"), want: true},
		{name: "unique constraint message", err: errors.New("UNIQUE constraint failed: revoked_tokens.jti"), want: true},
		{name: "other error", err: errors.New("network timeout"), want: false},
		{name: "nil error", err: nil, want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := repository.IsDuplicateRevokedTokenError(tc.err); got != tc.want {
				t.Fatalf("IsDuplicateRevokedTokenError() = %v, want %v", got, tc.want)
			}
		})
	}
}
