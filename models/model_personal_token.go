package models

import (
	"gorm.io/gorm"
)

type PersonalAccessTokens struct {
	gorm.Model
	Token string `gorm:"type:varchar(255);not null" json:"token"`
}