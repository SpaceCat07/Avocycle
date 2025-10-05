package models

import (
	"gorm.io/gorm"
)

type ProsesProduksi struct {
	gorm.Model
	Fase string `gorm:"type:varchar(50);not null" json:"fase"`
}