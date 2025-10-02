package models

import (
	"gorm.io/gorm"
)

type Kebun struct {
	gorm.Model
	NamaKebun string `gorm:"type:varchar(100);not null" json:"nama_kebun"`
	MDPL string `gorm:"type:varchar(100);not null" json:"mdpl"`
}