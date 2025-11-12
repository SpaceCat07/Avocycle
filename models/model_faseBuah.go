package models

import (
	"time"

	"gorm.io/gorm"
)

type FaseBuah struct {
	gorm.Model
	MingguKe      	int       	`gorm:"not null" json:"minggu_ke"`
	TanggalCatat  	*time.Time 	`gorm:"not null" json:"tanggal_catat"`
	TanggalCover  	*time.Time 	`gorm:"not null" json:"tanggal_cover"`
	JumlahCover   	int       	`gorm:"not null" json:"jumlah_cover"`
	WarnaLabel    	string    	`gorm:"size:50" json:"warna_label,omitempty"`
	EstimasiPanen 	*time.Time 	`gorm:"not null;index" json:"estimasi_panen"`
	TanamanID 		uint    	`gorm:"not null;index" json:"tanaman_id"`
	Tanaman   		Tanaman 	`gorm:"foreignKey:TanamanID;references:ID" json:"tanaman"`
}