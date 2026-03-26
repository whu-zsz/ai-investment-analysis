package model

import (
	"time"
)

type UploadedFile struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint64     `gorm:"not null;index" json:"user_id"`
	FileName        string     `gorm:"size:255;not null" json:"file_name"`
	FilePath        string     `gorm:"size:500;not null" json:"file_path"`
	FileSize        int64      `gorm:"not null" json:"file_size"`
	FileType        string     `gorm:"size:10;not null" json:"file_type"` // csv, xlsx, xls
	UploadStatus    string     `gorm:"size:20;default:'pending'" json:"upload_status"` // pending, processing, success, failed
	RecordsImported int        `gorm:"default:0" json:"records_imported"`
	ErrorMessage    *string    `gorm:"type:text" json:"error_message"`
	UploadedAt      time.Time  `json:"uploaded_at"`
	ProcessedAt     *time.Time `json:"processed_at"`
}

func (UploadedFile) TableName() string {
	return "uploaded_files"
}
