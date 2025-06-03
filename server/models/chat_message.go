package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	FamilyID  uint           `gorm:"index;not null" json:"family_id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Content   string         `gorm:"type:text" json:"content,omitempty"`
	MediaURL  *string        `json:"media_url,omitempty"`
	ReplyToID *uint          `json:"reply_to_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
