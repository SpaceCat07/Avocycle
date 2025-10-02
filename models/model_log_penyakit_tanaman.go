package models

import (
	"gorm.io/gorm"
)

type LogPenyakitTanaman struct {
	gorm.Model
	Kondisi    string `gorm:"type:enum('Parah','Sedang','Ringan','Sembuh');not null" json:"kondisi"`
	Catatan    string `gorm:"type:text" json:"catatan"`
	Foto       string `gorm:"type:varchar(255)" json:"foto"`
	TanamanID  uint   `gorm:"not null;index" json:"tanaman_id"`
	Tanaman    Tanaman `gorm:"foreignKey:TanamanID;references:ID" json:"tanaman"`
	PenyakitID uint   `gorm:"not null;index" json:"penyakit_id"`
	Penyakit   PenyakitTanaman `gorm:"foreignKey:PenyakitID;references:ID" json:"penyakit"`
}