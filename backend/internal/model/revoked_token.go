package model

import "time"

type RevokedToken struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint64    `gorm:"not null;index" json:"user_id"`
	JTI            string    `gorm:"size:64;not null" json:"jti"`
	TokenExpiresAt time.Time `gorm:"not null;index" json:"token_expires_at"`
	RevokedAt      time.Time `gorm:"not null" json:"revoked_at"`
	Reason         string    `gorm:"size:20;not null;default:'logout'" json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
}

func (RevokedToken) TableName() string {
	return "revoked_tokens"
}
