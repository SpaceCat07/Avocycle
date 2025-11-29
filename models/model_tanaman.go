package models

import (
	"gorm.io/gorm"
	"time"
)

type Tanaman struct {
	gorm.Model
	NamaTanaman  string    `gorm:"type:varchar(100);not null" json:"nama_tanaman"`
	Varietas     string    `gorm:"type:varchar(50);check:varietas IN ('Var1', 'Var2', 'Var3')" json:"varietas"`
	TanggalTanam time.Time `gorm:"type:date;not null" json:"tanggal_tanam"`
	KebunID      uint      `gorm:"not null;index" json:"kebun_id"`
	Kebun        Kebun     `gorm:"foreignKey:KebunID;references:ID" json:"kebun"`
	KodeBlok	 string    `gorm:"type:varchar(25);not null" json:"kode_blok"`
	KodeTanaman  string    `gorm:"type:varchar(50);not null" json:"kode_tanaman"`
	FotoTanaman  string    `gorm:"type:text" json:"foto_tanaman,omitempty"`
	MasaProduksi int	   `gorm:"not null" json:"masa_produksi"`
	FotoTanamanID string   `gorm:"type:varchar(255)" json:"foto_tanaman_id,omitempty"`
}