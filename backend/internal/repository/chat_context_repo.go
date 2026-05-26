package repository

import (
	"strings"

	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type ChatContextRepository interface {
	Create(entity *model.ChatContext) error
	Update(entity *model.ChatContext) error
	FindByID(id uint64) (*model.ChatContext, error)
	FindLatestByUserReport(userID, reportID uint64, contextType string) (*model.ChatContext, error)
}

type chatContextRepository struct {
	db *gorm.DB
}

func NewChatContextRepository(db *gorm.DB) ChatContextRepository {
	return &chatContextRepository{db: db}
}

func (r *chatContextRepository) Create(entity *model.ChatContext) error {
	return r.db.Create(entity).Error
}

func (r *chatContextRepository) Update(entity *model.ChatContext) error {
	return r.db.Save(entity).Error
}

func (r *chatContextRepository) FindByID(id uint64) (*model.ChatContext, error) {
	var entity model.ChatContext
	if err := r.db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *chatContextRepository) FindLatestByUserReport(userID, reportID uint64, contextType string) (*model.ChatContext, error) {
	var entity model.ChatContext
	query := r.db.Where("user_id = ? AND report_id = ?", userID, reportID)
	if contextType = strings.TrimSpace(contextType); contextType != "" {
		query = query.Where("context_type = ?", contextType)
	}
	if err := query.Order("updated_at DESC").First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}
