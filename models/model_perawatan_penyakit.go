package models

import (
	"gorm.io/gorm"
)

type PerawatanPenyakit struct {
	gorm.Model
	Tindakan             string `gorm:"type:text;not null" json:"tindakan"`
	LogPenyakitTanamanID uint   `gorm:"not null;index" json:"log_penyakit_tanaman_id"`
	LogPenyakitTanaman   LogPenyakitTanaman `gorm:"foreignKey:LogPenyakitTanamanID;references:ID" json:"log_penyakit_tanaman"`
}