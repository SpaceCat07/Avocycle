package models

import (
	"gorm.io/gorm"
)

type ProsesProduksi struct {
	gorm.Model
	Fase string `gorm:"type:varchar(20);check:fase IN ('Berbunga', 'Berbuah', 'Panen')" json:"fase"`
}