package models

import (
	"time"

	"gorm.io/gorm"
)

type FamilySubscription struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FamilyID  uint      `json:"family_id"`
	IsActive  bool      `gorm:"default:false" json:"is_active"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
