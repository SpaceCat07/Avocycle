package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PetaniDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Selamat datang Petani!",
	})
}
