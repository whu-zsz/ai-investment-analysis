package repository

import (
	"time"

	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type MarketSnapshotRepository interface {
	BatchCreate(snapshots []model.MarketSnapshot) error
	FindLatestBatchNo() (string, error)
	FindByBatchNo(batchNo string) ([]model.MarketSnapshot, error)
	FindHistoryBySymbol(symbol string, limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error)
}

type marketSnapshotRepository struct {
	db *gorm.DB
}

func NewMarketSnapshotRepository(db *gorm.DB) MarketSnapshotRepository {
	return &marketSnapshotRepository{db: db}
}

func (r *marketSnapshotRepository) BatchCreate(snapshots []model.MarketSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return r.db.CreateInBatches(snapshots, 100).Error
}

func (r *marketSnapshotRepository) FindLatestBatchNo() (string, error) {
	var snapshot model.MarketSnapshot
	err := r.db.Order("snapshot_time DESC, id DESC").First(&snapshot).Error
	if err != nil {
		return "", err
	}
	return snapshot.BatchNo, nil
}

func (r *marketSnapshotRepository) FindByBatchNo(batchNo string) ([]model.MarketSnapshot, error) {
	var snapshots []model.MarketSnapshot
	err := r.db.Where("batch_no = ?", batchNo).Order("symbol ASC").Find(&snapshots).Error
	return snapshots, err
}

func (r *marketSnapshotRepository) FindHistoryBySymbol(symbol string, limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	if limit <= 0 {
		limit = 60
	}

	var snapshots []model.MarketSnapshot
	db := r.db.Where("symbol = ?", symbol)
	if startTime != nil {
		db = db.Where("snapshot_time >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("snapshot_time <= ?", *endTime)
	}

	err := db.Order("snapshot_time DESC").Limit(limit).Find(&snapshots).Error
	return snapshots, err
}
