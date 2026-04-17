package repository

import (
	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type AnalysisReportItemRepository interface {
	BatchCreate(items []model.AnalysisReportItem) error
	FindByReportID(reportID uint64) ([]model.AnalysisReportItem, error)
}

type analysisReportItemRepository struct {
	db *gorm.DB
}

func NewAnalysisReportItemRepository(db *gorm.DB) AnalysisReportItemRepository {
	return &analysisReportItemRepository{db: db}
}

func (r *analysisReportItemRepository) BatchCreate(items []model.AnalysisReportItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.CreateInBatches(items, 100).Error
}

func (r *analysisReportItemRepository) FindByReportID(reportID uint64) ([]model.AnalysisReportItem, error) {
	var items []model.AnalysisReportItem
	err := r.db.Where("report_id = ?", reportID).Order("symbol ASC").Find(&items).Error
	return items, err
}
