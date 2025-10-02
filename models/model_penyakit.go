package models

import (
	"gorm.io/gorm"
)

type PenyakitTanaman struct {
	gorm.Model
	NamaPenyakit string `gorm:"type:varchar(100);not null" json:"nama_penyakit"`
	Deskripsi string `gorm:"type:varchar(255);not null" json:"deskripsi"`
}