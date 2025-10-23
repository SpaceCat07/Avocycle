package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CREATE: Tambah Kebun
func CreateKebun(c *gin.Context) {
	var kebun models.Kebun

	if err := c.ShouldBindJSON(&kebun); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal konek ke database"})
		return
	}

	if err := db.Create(&kebun).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kebun berhasil dibuat",
		"data":    kebun,
	})
}

// READ ALL: Ambil semua Kebun
func GetAllKebun(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal konek ke database"})
		return
	}

	var kebuns []models.Kebun
	if err := db.Find(&kebuns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kebuns,
	})
}

// READ ONE: Ambil Kebun by ID
func GetKebunByID(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal konek ke database"})
		return
	}

	var kebun models.Kebun
	if err := db.First(&kebun, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Kebun tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": kebun})
}

// UPDATE: Edit Kebun
func UpdateKebun(c *gin.Context) {
	id := c.Param("id")

	var updateData models.Kebun
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal konek ke database"})
		return
	}

	var kebun models.Kebun
	if err := db.First(&kebun, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Kebun tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	kebun.NamaKebun = updateData.NamaKebun
	kebun.MDPL = updateData.MDPL

	if err := db.Save(&kebun).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kebun berhasil diperbarui",
		"data":    kebun,
	})
}

// DELETE: Hapus Kebun
func DeleteKebun(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal konek ke database"})
		return
	}

	var kebun models.Kebun
	if err := db.First(&kebun, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Kebun tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.Delete(&kebun).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kebun berhasil dihapus"})
}
