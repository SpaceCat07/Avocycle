package models

import (
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    FullName     string `gorm:"type:varchar(100);not null" json:"fullname"`
    Phone        string `gorm:"type:varchar(50);not null" json:"phone"`
    Email        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"email"`
    PasswordHash string `gorm:"size:255" json:"password"`
    AuthProvider string `gorm:"type:varchar(20);check:auth_provider IN ('Google', 'Local');" json:"auth_provider"`
    ProviderID   string `gorm:"size:255" json:"provider_id"`
    Role         string `gorm:"type:varchar(20);check:role IN ('Admin', 'Petani', 'Pembeli');" json:"role"`
}