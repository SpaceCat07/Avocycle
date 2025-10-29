package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"Avocycle/config"
	"Avocycle/models"
)

func GetAllTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pageParam := c.DefaultQuery("page", "1")
	limitParam := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := db.Model(&models.Tanaman{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var tanaman []models.Tanaman
	if err := db.Preload("Kebun").Limit(limit).Offset(offset).Find(&tanaman).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data": tanaman,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func GetTanamanByID(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	var tanaman models.Tanaman
	if err := db.Preload("Kebun").First(&tanaman, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tanaman tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tanaman)
}

func CreateTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		NamaTanaman  string `json:"nama_tanaman" binding:"required"`
		Varietas     string `json:"varietas" binding:"required"`
		TanggalTanam string `json:"tanggal_tanam" binding:"required"`
		KebunID      uint   `json:"kebun_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// parse
	parsedTime, err := time.Parse("2006-01-02", input.TanggalTanam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}

	tanaman := models.Tanaman{
		NamaTanaman:  input.NamaTanaman,
		Varietas:     input.Varietas,
		TanggalTanam: parsedTime,
		KebunID:      input.KebunID,
	}

	if err := db.Create(&tanaman).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tanaman)
}

func UpdateTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	var tanaman models.Tanaman
	if err := db.First(&tanaman, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tanaman tidak ditemukan"})
		return
	}

	var input struct {
		NamaTanaman  string    `json:"nama_tanaman"`
		Varietas     string    `json:"varietas"`
		TanggalTanam string `json:"tanggal_tanam"`
		KebunID      uint      `json:"kebun_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// parse
	parsedTime, err := time.Parse("2006-01-02", input.TanggalTanam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}

	tanaman.NamaTanaman = input.NamaTanaman
	tanaman.Varietas = input.Varietas
	tanaman.TanggalTanam = parsedTime
	tanaman.KebunID = input.KebunID

	if err := db.Save(&tanaman).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tanaman)
}

func DeleteTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if err := db.Delete(&models.Tanaman{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tanaman berhasil dihapus"})
}
