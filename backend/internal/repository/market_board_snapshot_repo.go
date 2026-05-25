package repository

import (
	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type MarketBoardSnapshotRepository interface {
	BatchCreate(snapshots []model.MarketBoardSnapshot) error
	FindLatestBatchNo(boardType string) (string, error)
	FindByBatchNo(boardType, batchNo string, limit int) ([]model.MarketBoardSnapshot, error)
	FindLatest(limit int) ([]model.MarketBoardSnapshot, error)
}

type marketBoardSnapshotRepository struct {
	db *gorm.DB
}

func NewMarketBoardSnapshotRepository(db *gorm.DB) MarketBoardSnapshotRepository {
	return &marketBoardSnapshotRepository{db: db}
}

func (r *marketBoardSnapshotRepository) BatchCreate(snapshots []model.MarketBoardSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return r.db.CreateInBatches(snapshots, 100).Error
}

func (r *marketBoardSnapshotRepository) FindLatestBatchNo(boardType string) (string, error) {
	var snapshot model.MarketBoardSnapshot
	err := r.db.Where("board_type = ?", boardType).Order("created_at DESC, id DESC").First(&snapshot).Error
	if err != nil {
		return "", err
	}
	return snapshot.BatchNo, nil
}

func (r *marketBoardSnapshotRepository) FindByBatchNo(boardType, batchNo string, limit int) ([]model.MarketBoardSnapshot, error) {
	if limit <= 0 {
		limit = 20
	}
	var snapshots []model.MarketBoardSnapshot
	err := r.db.Where("board_type = ? AND batch_no = ?", boardType, batchNo).
		Order("change_percent DESC, turnover DESC").
		Limit(limit).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *marketBoardSnapshotRepository) FindLatest(limit int) ([]model.MarketBoardSnapshot, error) {
	if limit <= 0 {
		limit = 40
	}
	var snapshots []model.MarketBoardSnapshot
	err := r.db.Raw(`
		SELECT mbs.*
		FROM market_board_snapshots mbs
		JOIN (
			SELECT board_type, MAX(created_at) AS max_created_at
			FROM market_board_snapshots
			GROUP BY board_type
		) latest ON latest.board_type = mbs.board_type AND latest.max_created_at = mbs.created_at
		ORDER BY mbs.board_type ASC, mbs.change_percent DESC, mbs.turnover DESC
		LIMIT ?
	`, limit).Scan(&snapshots).Error
	return snapshots, err
}
