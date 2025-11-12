package models

import (
	"time"

	"gorm.io/gorm"
)

type FaseBunga struct {
	gorm.Model
	MingguKe		int			`gorm:"not null" json:"minggu_ke"`
	TanggalCatat	*time.Time	`gorm:"not null" json:"tanggal_catat"`
	JumlahBunga		int			`gorm:"default:0" json:"jumlah_bunga"`
	BungaPecah		int			`gorm:"default:0" json:"bunga_pecah"`
	PentilMuncul	int			`gorm:"default:0" json:"pentil_muncul"`
	TanamanID 		uint    	`gorm:"not null;index" json:"tanaman_id"`
	Tanaman   		Tanaman 	`gorm:"foreignKey:TanamanID;references:ID" json:"tanaman"`
}