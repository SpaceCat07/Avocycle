package models

import (
	"gorm.io/gorm"
)

type Buah struct {
	gorm.Model
	NamaBuah string `gorm:"type:varchar(100);not null" json:"nama_buah"`
	TanamanID uint `gorm:"not null" json:"tanaman_id"`
	Tanaman Tanaman `gorm:"foreginKey:TanamanID; references:ID" json:"tanaman"`
}