package repository

import (
	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockQuoteDetailRepository interface {
	Upsert(detail *model.StockQuoteDetail) error
	FindBySymbol(symbol string) (*model.StockQuoteDetail, error)
	FindBySymbols(symbols []string) ([]model.StockQuoteDetail, error)
}

type stockQuoteDetailRepository struct {
	db *gorm.DB
}

func NewStockQuoteDetailRepository(db *gorm.DB) StockQuoteDetailRepository {
	return &stockQuoteDetailRepository{db: db}
}

func (r *stockQuoteDetailRepository) Upsert(detail *model.StockQuoteDetail) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "market", "last_price", "open_price", "high_price", "low_price", "prev_close",
			"change_amount", "change_percent", "volume", "turnover", "volume_ratio", "turnover_rate",
			"amplitude", "limit_up", "limit_down", "average_price", "total_shares", "float_shares",
			"total_market_cap", "float_market_cap", "industry", "region", "concepts", "source", "fetched_at", "updated_at",
		}),
	}).Create(detail).Error
}

func (r *stockQuoteDetailRepository) FindBySymbol(symbol string) (*model.StockQuoteDetail, error) {
	var detail model.StockQuoteDetail
	if err := r.db.Where("symbol = ?", symbol).First(&detail).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *stockQuoteDetailRepository) FindBySymbols(symbols []string) ([]model.StockQuoteDetail, error) {
	if len(symbols) == 0 {
		return []model.StockQuoteDetail{}, nil
	}

	var details []model.StockQuoteDetail
	err := r.db.Where("symbol IN ?", symbols).Find(&details).Error
	return details, err
}
