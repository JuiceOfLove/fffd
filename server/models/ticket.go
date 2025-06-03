package models

import (
	"time"

	"gorm.io/gorm"
)

type Ticket struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	OperatorID    *uint          `json:"operator_id,omitempty"`
	Subject       string         `gorm:"size:255;not null" json:"subject"`
	Status        string         `gorm:"size:20;not null" json:"status"`
	LastMessageAt time.Time      `gorm:"not null" json:"last_message_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
