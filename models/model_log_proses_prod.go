package models

import (
	"gorm.io/gorm"
	"time"
)

type LogProsesProduksi struct {
	gorm.Model
	Deskripsi     string    `gorm:"type:text;not null" json:"deskripsi"`
	PrediksiPanen time.Time `gorm:"type:date;not null" json:"prediksi_panen"`
	ProsesID      uint      `gorm:"not null;index" json:"proses_id"`
	Proses        ProsesProduksi `gorm:"foreignKey:ProsesID;references:ID" json:"proses_produksi"`
	BuahID        uint      `gorm:"not null;index" json:"buah_id"`
	Buah          Buah      `gorm:"foreignKey:BuahID;references:ID" json:"buah"`
}