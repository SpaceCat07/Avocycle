package models

import (
	"time"

	"gorm.io/gorm"
)

type FasePanen struct {
	gorm.Model
	TanggalPanenAktual 	*time.Time	`gorm:"index" json:"tanggal_panen_aktual,omitempty"`     // Tanggal panen sebenarnya
	JumlahPanen        	int        	`gorm:"default:0" json:"jumlah_panen"`                   // Jumlah buah yang dipanen
	JumlahSampel       	int        	`gorm:"default:0" json:"jumlah_sampel"`                  // Jumlah buah untuk sampel
	BeratTotal         	float64    	`gorm:"type:decimal(10,2);default:0" json:"berat_total"` // Berat total (kg)
	Catatan            	string     	`gorm:"type:text" json:"catatan,omitempty"`              // Catatan panen
	FotoPanen          	string     	`gorm:"type:text" json:"foto_panen,omitempty"`           // Foto hasil panen (optional)
	FotoPanenID			string		`gorm:"type:varchar(255)" json:"foto_panen_id,omitempty"`
	TanamanID 			uint    	`gorm:"not null;index" json:"tanaman_id"`
	Tanaman   			Tanaman 	`gorm:"foreignKey:TanamanID;references:ID" json:"tanaman"`
}