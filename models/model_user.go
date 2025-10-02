package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FullName string `gorm:"type:varchar(100);not null" json:"fullname"`
	Phone string `gorm:"type:varchar(50);not null" json:"phone"`
	Email string `gorm:"type:varchar(100);not null;uniqueIndex" json:"email"`
	PasswordHash string `gorm:"size:255" json:"password"`
	AuthProvider string `gorm:"enum('Google', 'Local');" json:"auth_provider"`
	ProviderID string `gorm:"size:255" json:"provider_id"`
	Role string `gorm:"enum('Admin', 'Petani', 'Pembeli');" json:"Role"`
}