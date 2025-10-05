package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	FullName           string    `gorm:"type:varchar(100);not null" json:"fullname"`
	Phone              string    `gorm:"type:varchar(20);not null" json:"phone"`
	Email              string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash       string    `gorm:"type:varchar(255);not null" json:"-"`
	AuthProvider	   string 	 `gorm:"type:varchar(20);default:'Local';not null"`
	ProviderID         string    `gorm:"type:varchar(255)" json:"provider_id,omitempty"`
	Role               string    `gorm:"type:varchar(20);default:'Petani';not null" json:"role"`
	FailedLoginAttempts int       `gorm:"default:0" json:"-"`
	LastLoginAttempt   time.Time `json:"-"` // untuk rate limiting
}
