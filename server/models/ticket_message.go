package models

import (
	"time"

	"gorm.io/gorm"
)

type TicketMessage struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	TicketID   uint           `gorm:"index;not null" json:"ticket_id"`
	SenderID   uint           `gorm:"not null" json:"sender_id"`
	SenderRole string         `gorm:"size:20;not null" json:"sender_role"`
	Content    string         `gorm:"type:text" json:"content"`
	MediaURL   *string        `json:"media_url,omitempty"`
	ReplyToID  *uint          `json:"reply_to_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}