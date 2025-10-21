package config

import (
	"gopkg.in/danilopolani/gocialite.v1"
)

var Gocial *gocialite.Dispatcher

func InitGocial() {
	Gocial = gocialite.NewDispatcher()
}