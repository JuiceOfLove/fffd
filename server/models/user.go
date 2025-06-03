package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Email          string         `gorm:"unique;not null" json:"email"`
	Password       string         `gorm:"not null" json:"-"` // пароль не возвращается в JSON
	Role           string         `gorm:"size:50;default:'user'" json:"role"`
	FamilyID       uint           `json:"family_id"` // Если 0, то семья не создана
	IsActivated    bool           `gorm:"default:false" json:"isActivated"`
	ActivationLink string         `json:"activationLink"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}