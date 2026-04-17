package model

import "time"

type AnalysisTask struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID              uint64     `gorm:"not null;index" json:"user_id"`
	TaskType            string     `gorm:"size:50;not null" json:"task_type"`
	Status              string     `gorm:"size:20;not null;default:'pending';index" json:"status"`
	ProgressStage       string     `gorm:"size:50;not null;default:'pending'" json:"progress_stage"`
	AnalysisPeriodStart time.Time  `gorm:"not null;type:date" json:"analysis_period_start"`
	AnalysisPeriodEnd   time.Time  `gorm:"not null;type:date" json:"analysis_period_end"`
	RequestPayload      *string    `gorm:"type:json" json:"request_payload"`
	ResultReportID      *uint64    `gorm:"index" json:"result_report_id"`
	ErrorMessage        *string    `gorm:"type:text" json:"error_message"`
	StartedAt           *time.Time `json:"started_at"`
	FinishedAt          *time.Time `json:"finished_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (AnalysisTask) TableName() string {
	return "ai_analysis_tasks"
}
