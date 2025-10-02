package models

import (
	"gorm.io/gorm"
	"time"
)

type Tanaman struct {
	gorm.Model
	NamaTanaman string `gorm:"type:varchar(100);not null" json:"nama_tanaman"`
	Varietas string `gorm:"enum('Var1', 'Var2', 'Var3);" json:"varietas"`
	TanggalTanam time.Time `gorm:"type:date;not null" json:"tanggal_tanam"`
	KebunID uint `gorm:"not null" json:"kebun_id"`
	Kebun Kebun `gorm:"foreginKey:KebunID; references:ID" json:"kebun"`
}