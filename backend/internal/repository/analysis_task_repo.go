package repository

import (
	"time"

	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type AnalysisTaskRepository interface {
	Create(task *model.AnalysisTask) error
	FindByIDAndUserID(id, userID uint64) (*model.AnalysisTask, error)
	FindByUserID(userID uint64, status string, limit, offset int) ([]model.AnalysisTask, int64, error)
	HasRunningTask(userID uint64, taskType string) (bool, error)
	UpdateProgress(id uint64, status, stage string, errorMsg *string, resultReportID *uint64, startedAt, finishedAt *time.Time) error
}

type analysisTaskRepository struct {
	db *gorm.DB
}

func NewAnalysisTaskRepository(db *gorm.DB) AnalysisTaskRepository {
	return &analysisTaskRepository{db: db}
}

func (r *analysisTaskRepository) Create(task *model.AnalysisTask) error {
	return r.db.Create(task).Error
}

func (r *analysisTaskRepository) FindByIDAndUserID(id, userID uint64) (*model.AnalysisTask, error) {
	var task model.AnalysisTask
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *analysisTaskRepository) FindByUserID(userID uint64, status string, limit, offset int) ([]model.AnalysisTask, int64, error) {
	var tasks []model.AnalysisTask
	var total int64

	query := r.db.Model(&model.AnalysisTask{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Order("created_at DESC, id DESC").Find(&tasks).Error
	return tasks, total, err
}

func (r *analysisTaskRepository) HasRunningTask(userID uint64, taskType string) (bool, error) {
	var count int64
	query := r.db.Model(&model.AnalysisTask{}).Where("user_id = ? AND status IN ?", userID, []string{"pending", "processing"})
	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *analysisTaskRepository) UpdateProgress(id uint64, status, stage string, errorMsg *string, resultReportID *uint64, startedAt, finishedAt *time.Time) error {
	updates := map[string]interface{}{}
	if status != "" {
		updates["status"] = status
	}
	if stage != "" {
		updates["progress_stage"] = stage
	}
	if errorMsg != nil {
		updates["error_message"] = *errorMsg
	}
	if resultReportID != nil {
		updates["result_report_id"] = *resultReportID
	}
	if startedAt != nil {
		updates["started_at"] = *startedAt
	}
	if finishedAt != nil {
		updates["finished_at"] = *finishedAt
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.AnalysisTask{}).Where("id = ?", id).Updates(updates).Error
}
