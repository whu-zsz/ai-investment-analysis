package repository

import (
	"stock-analysis-backend/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PortfolioRepository interface {
	Create(portfolio *model.Portfolio) error
	FindByID(id uint64) (*model.Portfolio, error)
	FindByUserID(userID uint64) ([]model.Portfolio, error)
	FindByUserAndAsset(userID uint64, assetCode string) (*model.Portfolio, error)
	Update(portfolio *model.Portfolio) error
	Delete(id uint64) error
	UpdateCurrentPrice(assetCode string, currentPrice decimal.Decimal) error
}

type portfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) PortfolioRepository {
	return &portfolioRepository{db: db}
}

func (r *portfolioRepository) Create(portfolio *model.Portfolio) error {
	return r.db.Create(portfolio).Error
}

func (r *portfolioRepository) FindByID(id uint64) (*model.Portfolio, error) {
	var portfolio model.Portfolio
	err := r.db.First(&portfolio, id).Error
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (r *portfolioRepository) FindByUserID(userID uint64) ([]model.Portfolio, error) {
	var portfolios []model.Portfolio
	err := r.db.Where("user_id = ?", userID).Order("last_updated DESC").Find(&portfolios).Error
	return portfolios, err
}

func (r *portfolioRepository) FindByUserAndAsset(userID uint64, assetCode string) (*model.Portfolio, error) {
	var portfolio model.Portfolio
	err := r.db.Where("user_id = ? AND asset_code = ?", userID, assetCode).First(&portfolio).Error
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (r *portfolioRepository) Update(portfolio *model.Portfolio) error {
	return r.db.Save(portfolio).Error
}

func (r *portfolioRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Portfolio{}, id).Error
}

func (r *portfolioRepository) UpdateCurrentPrice(assetCode string, currentPrice decimal.Decimal) error {
	return r.db.Model(&model.Portfolio{}).
		Where("asset_code = ?", assetCode).
		Update("current_price", currentPrice).Error
}
