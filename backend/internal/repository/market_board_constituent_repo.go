package repository

import (
	"time"

	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type MarketBoardConstituentRepository interface {
	ReplaceAll(items []model.MarketBoardConstituent) error
	FindAll() ([]model.MarketBoardConstituent, error)
	FindLatestSyncedAt() (time.Time, error)
}

type marketBoardConstituentRepository struct {
	db *gorm.DB
}

func NewMarketBoardConstituentRepository(db *gorm.DB) MarketBoardConstituentRepository {
	return &marketBoardConstituentRepository{db: db}
}

func (r *marketBoardConstituentRepository) ReplaceAll(items []model.MarketBoardConstituent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.MarketBoardConstituent{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.CreateInBatches(items, 500).Error
	})
}

func (r *marketBoardConstituentRepository) FindAll() ([]model.MarketBoardConstituent, error) {
	var items []model.MarketBoardConstituent
	err := r.db.Order("board_type ASC, board_code ASC, symbol ASC").Find(&items).Error
	return items, err
}

func (r *marketBoardConstituentRepository) FindLatestSyncedAt() (time.Time, error) {
	var item model.MarketBoardConstituent
	err := r.db.Order("synced_at DESC, id DESC").First(&item).Error
	if err != nil {
		return time.Time{}, err
	}
	return item.SyncedAt, nil
}
