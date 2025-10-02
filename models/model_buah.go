package models

import (
	"gorm.io/gorm"
)

type Buah struct {
	gorm.Model
	NamaBuah  string  `gorm:"type:varchar(100);not null" json:"nama_buah"`
	TanamanID uint    `gorm:"not null;index" json:"tanaman_id"`
	Tanaman   Tanaman `gorm:"foreignKey:TanamanID;references:ID" json:"tanaman"`
}