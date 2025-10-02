package models

import (
	"gorm.io/gorm"
)

type ProsesProduksi struct {
	gorm.Model
	Fase string `gorm:"enum('Berbunga', 'Berbuah', 'Panen');" json:"fase"`
}