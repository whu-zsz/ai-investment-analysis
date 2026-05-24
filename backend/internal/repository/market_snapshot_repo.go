package repository

import (
	"strings"
	"time"

	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type MarketSnapshotRepository interface {
	BatchCreate(snapshots []model.MarketSnapshot) error
	FindLatestBatchNo() (string, error)
	FindByBatchNo(batchNo string) ([]model.MarketSnapshot, error)
	FindLatestBySymbol(symbol string) (*model.MarketSnapshot, error)
	FindHistory(limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error)
	FindHistoryBySymbol(symbol string, limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error)
	SearchStocks(query string, limit int) ([]model.MarketSnapshot, error)
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
	err := r.db.Order("created_at DESC, id DESC").First(&snapshot).Error
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

func (r *marketSnapshotRepository) FindLatestBySymbol(symbol string) (*model.MarketSnapshot, error) {
	var snapshot model.MarketSnapshot
	err := r.db.Where("symbol = ?", symbol).Order("snapshot_time DESC, id DESC").First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *marketSnapshotRepository) FindHistory(limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	if limit <= 0 {
		limit = 60
	}

	var snapshots []model.MarketSnapshot
	db := r.db.Model(&model.MarketSnapshot{})
	if startTime != nil {
		db = db.Where("snapshot_time >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("snapshot_time <= ?", *endTime)
	}

	err := db.Order("snapshot_time DESC, id DESC").Limit(limit).Find(&snapshots).Error
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

	err := db.Order("snapshot_time DESC, id DESC").Limit(limit).Find(&snapshots).Error
	return snapshots, err
}

func (r *marketSnapshotRepository) SearchStocks(query string, limit int) ([]model.MarketSnapshot, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query = strings.TrimSpace(query)
	like := "%" + query + "%"
	var snapshots []model.MarketSnapshot
	err := r.db.Raw(`
		SELECT ms.*
		FROM market_snapshots ms
		JOIN (
			SELECT symbol, MAX(snapshot_time) AS max_time
			FROM market_snapshots
			WHERE symbol LIKE ? OR name LIKE ?
			GROUP BY symbol
			LIMIT ?
		) latest ON latest.symbol = ms.symbol AND latest.max_time = ms.snapshot_time
		ORDER BY CASE WHEN ms.symbol = ? THEN 0 WHEN ms.symbol LIKE ? THEN 1 ELSE 2 END, ms.symbol ASC, ms.id DESC
	`, like, like, limit, query, like).Scan(&snapshots).Error
	return snapshots, err
}
