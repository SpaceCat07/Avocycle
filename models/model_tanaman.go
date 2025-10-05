package models

import (
	"gorm.io/gorm"
	"time"
)

type Tanaman struct {
	gorm.Model
	NamaTanaman string `gorm:"type:varchar(100);not null" json:"nama_tanaman"`
	Varietas    string `gorm:"type:varchar(50);not null" json:"varietas"`
	TanggalTanam time.Time `gorm:"type:date;not null" json:"tanggal_tanam"`
	KebunID uint `gorm:"not null" json:"kebun_id"`
	Kebun Kebun `gorm:"foreignKey:KebunID; references:ID" json:"kebun"`
}

func (Tanaman) TableName() string {
	return "tanamans"
}