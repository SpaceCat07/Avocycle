package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
)

func CountAllPohon(c *gin.Context) {
	// connect to db
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var totalTree int64
	if err := db.Model(&models.Tanaman{}).Count(&totalTree).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count tanaman", err.Error())
		return
	}

	if totalTree == 0 {
		utils.SuccessResponse(c, http.StatusOK, "No tanaman data", totalTree)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Count tanaman data", totalTree)
}

func CountTanamanDiseased(c *gin.Context) {
	// connect to db
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var tanamanSakit int64

	// Gunakan DISTINCT untuk count tanaman_id unik
	if err := db.Model(&models.LogPenyakitTanaman{}).
		Select("DISTINCT tanaman_id").
		Where("kondisi IN ?", []string{"Parah", "Sedang", "Ringan"}).
		Where("created_at IN (?)",
			db.Model(&models.LogPenyakitTanaman{}).
				Select("MAX(created_at)").
				Where("tanaman_id = log_penyakit_tanaman.tanaman_id").
				Group("tanaman_id"),
		).
		Count(&tanamanSakit).
		Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count tanaman sakit", err.Error())
		return
	}

	if tanamanSakit == 0 {
		utils.SuccessResponse(c, http.StatusOK, "No tanaman sakit", tanamanSakit)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Count tanaman sakit data", tanamanSakit)
}

func CountSiapPanen(c *gin.Context) {
	// connect to db
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var siapPanen int64
	currentTime := time.Now()
	// Subquery untuk mendapatkan fase buah terbaru per tanaman
	subQuery := db.Model(&models.FaseBuah{}).
		Select("tanaman_id, MAX(created_at) as latest_created_at").
		Group("tanaman_id")

	// Query utama untuk count fase buah siap panen dari record terbaru
	if err := db.Model(&models.FaseBuah{}).
		Joins("INNER JOIN (?) as latest ON fase_buahs.tanaman_id = latest.tanaman_id AND fase_buahs.created_at = latest.latest_created_at", subQuery).
		Where("fase_buahs.estimasi_panen <= ?", currentTime).
		Count(&siapPanen).
		Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count tanaman siap panen", err.Error())
		return
	}

	if siapPanen == 0 {
		utils.SuccessResponse(c, http.StatusOK, "No tanaman siap panen", siapPanen)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Count tanaman siap panen", siapPanen)
}
