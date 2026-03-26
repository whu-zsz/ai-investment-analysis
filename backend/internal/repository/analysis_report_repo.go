package repository

import (
	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type AnalysisReportRepository interface {
	Create(report *model.AnalysisReport) error
	FindByID(id uint64) (*model.AnalysisReport, error)
	FindByUserID(userID uint64, reportType string, limit int) ([]model.AnalysisReport, error)
	FindLatestByUser(userID uint64, reportType string) (*model.AnalysisReport, error)
	Delete(id uint64) error
}

type analysisReportRepository struct {
	db *gorm.DB
}

func NewAnalysisReportRepository(db *gorm.DB) AnalysisReportRepository {
	return &analysisReportRepository{db: db}
}

func (r *analysisReportRepository) Create(report *model.AnalysisReport) error {
	return r.db.Create(report).Error
}

func (r *analysisReportRepository) FindByID(id uint64) (*model.AnalysisReport, error) {
	var report model.AnalysisReport
	err := r.db.First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *analysisReportRepository) FindByUserID(userID uint64, reportType string, limit int) ([]model.AnalysisReport, error) {
	var reports []model.AnalysisReport
	query := r.db.Where("user_id = ?", userID)

	if reportType != "" {
		query = query.Where("report_type = ?", reportType)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&reports).Error
	return reports, err
}

func (r *analysisReportRepository) FindLatestByUser(userID uint64, reportType string) (*model.AnalysisReport, error) {
	var report model.AnalysisReport
	err := r.db.Where("user_id = ? AND report_type = ?", userID, reportType).
		Order("created_at DESC").
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *analysisReportRepository) Delete(id uint64) error {
	return r.db.Delete(&model.AnalysisReport{}, id).Error
}
