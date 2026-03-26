package repository

import (
	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type UploadedFileRepository interface {
	Create(file *model.UploadedFile) error
	FindByID(id uint64) (*model.UploadedFile, error)
	FindByUserID(userID uint64) ([]model.UploadedFile, error)
	UpdateStatus(id uint64, status string, recordsImported int, errorMsg *string) error
}

type uploadedFileRepository struct {
	db *gorm.DB
}

func NewUploadedFileRepository(db *gorm.DB) UploadedFileRepository {
	return &uploadedFileRepository{db: db}
}

func (r *uploadedFileRepository) Create(file *model.UploadedFile) error {
	return r.db.Create(file).Error
}

func (r *uploadedFileRepository) FindByID(id uint64) (*model.UploadedFile, error) {
	var file model.UploadedFile
	err := r.db.First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *uploadedFileRepository) FindByUserID(userID uint64) ([]model.UploadedFile, error) {
	var files []model.UploadedFile
	err := r.db.Where("user_id = ?", userID).Order("uploaded_at DESC").Find(&files).Error
	return files, err
}

func (r *uploadedFileRepository) UpdateStatus(id uint64, status string, recordsImported int, errorMsg *string) error {
	updates := map[string]interface{}{
		"upload_status":     status,
		"records_imported":  recordsImported,
		"processed_at":      gorm.Expr("NOW()"),
	}
	if errorMsg != nil {
		updates["error_message"] = *errorMsg
	}
	return r.db.Model(&model.UploadedFile{}).Where("id = ?", id).Updates(updates).Error
}
