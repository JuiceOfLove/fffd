package models

import "time"

type FamilyInvitation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FamilyID  uint      `json:"family_id"`                    // семья, в которую приглашают
	Email     string    `gorm:"not null" json:"email"`        // email приглашённого
	Token     string    `gorm:"unique;not null" json:"token"` // уникальный токен приглашения
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
