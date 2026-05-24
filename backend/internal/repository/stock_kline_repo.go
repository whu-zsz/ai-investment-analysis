package repository

import (
	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockKlineRepository interface {
	UpsertBars(bars []model.StockKlineBar) error
	FindBars(symbol, period, adjust string, limit int) ([]model.StockKlineBar, error)
	FindLatestBar(symbol, period, adjust string) (*model.StockKlineBar, error)
}

type stockKlineRepository struct {
	db *gorm.DB
}

func NewStockKlineRepository(db *gorm.DB) StockKlineRepository {
	return &stockKlineRepository{db: db}
}

func (r *stockKlineRepository) UpsertBars(bars []model.StockKlineBar) error {
	if len(bars) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}, {Name: "period"}, {Name: "adjust_type"}, {Name: "bar_time"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"open_price", "close_price", "high_price", "low_price", "volume", "turnover", "amplitude",
			"change_percent", "change_amount", "turnover_rate", "source", "updated_at",
		}),
	}).Create(&bars).Error
}

func (r *stockKlineRepository) FindBars(symbol, period, adjust string, limit int) ([]model.StockKlineBar, error) {
	if limit <= 0 {
		limit = 120
	}
	var bars []model.StockKlineBar
	err := r.db.Where("symbol = ? AND period = ? AND adjust_type = ?", symbol, period, adjust).
		Order("bar_time DESC, id DESC").
		Limit(limit).
		Find(&bars).Error
	return bars, err
}

func (r *stockKlineRepository) FindLatestBar(symbol, period, adjust string) (*model.StockKlineBar, error) {
	var bar model.StockKlineBar
	err := r.db.Where("symbol = ? AND period = ? AND adjust_type = ?", symbol, period, adjust).
		Order("bar_time DESC, id DESC").
		First(&bar).Error
	if err != nil {
		return nil, err
	}
	return &bar, nil
}
