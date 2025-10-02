package models

import (
	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model
	UserID    uint    `gorm:"not null;index" json:"user_id"`
	User      User    `gorm:"foreignKey:UserID;references:ID" json:"user"`
	TanamanID uint    `gorm:"not null;index" json:"tanaman_id"`
	Tanaman   Tanaman `gorm:"foreignKey:TanamanID;references:ID" json:"tanaman"`
}