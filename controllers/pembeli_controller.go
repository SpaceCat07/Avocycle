package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PembeliDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Selamat datang Pembeli!",
	})
}
