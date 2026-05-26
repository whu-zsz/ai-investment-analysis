package model

import "time"

type ChatContext struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64     `gorm:"not null;index" json:"user_id"`
	ContextType string     `gorm:"size:32;not null;index" json:"context_type"`
	TargetKey    string     `gorm:"size:128;not null;index" json:"target_key"`
	Title       string     `gorm:"size:255;not null" json:"title"`
	ReportID    *uint64    `gorm:"index" json:"report_id"`
	MessagesJSON string    `gorm:"type:longtext;not null" json:"messages_json"`
	MetadataJSON string    `gorm:"type:longtext" json:"metadata_json"`
	LastQuestion string    `gorm:"size:1000" json:"last_question"`
	LastReply   string     `gorm:"type:longtext" json:"last_reply"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ChatContext) TableName() string {
	return "chat_contexts"
}
