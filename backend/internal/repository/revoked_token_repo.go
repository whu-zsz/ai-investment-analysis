package repository

import (
	"errors"
	"strings"

	"stock-analysis-backend/internal/model"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type RevokedTokenRepository interface {
	Create(token *model.RevokedToken) error
	ExistsByJTI(jti string) (bool, error)
}

type revokedTokenRepository struct {
	db *gorm.DB
}

func NewRevokedTokenRepository(db *gorm.DB) RevokedTokenRepository {
	return &revokedTokenRepository{db: db}
}

func (r *revokedTokenRepository) Create(token *model.RevokedToken) error {
	return r.db.Create(token).Error
}

func (r *revokedTokenRepository) ExistsByJTI(jti string) (bool, error) {
	jti = strings.TrimSpace(jti)
	if jti == "" {
		return false, nil
	}

	var revokedToken model.RevokedToken
	err := r.db.Select("jti").Where("jti = ?", jti).Limit(1).Take(&revokedToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func IsDuplicateRevokedTokenError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique constraint")
}
