package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Новый столбец CalendarID:
	CalendarID uint `json:"calendar_id"` // конкретный календарь, к которому относится событие

	FamilyID    uint      `json:"family_id"` // можно оставить, если хотите и семейную привязку
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"size:1000" json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedBy   uint      `json:"created_by"`
	IsCompleted bool      `gorm:"default:false" json:"is_completed"`
	Color       *string   `gorm:"size:20" json:"color,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
