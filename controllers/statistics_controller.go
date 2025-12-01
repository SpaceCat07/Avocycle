package controllers

import (
	"fmt"
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

func GetWeeklyPanenLast6Weeks(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -7*5) // 5 minggu ke belakang → total 6 minggu

	// struct lokal untuk nampung hasil Scan dari DB
	var rows []struct {
		Year       int
		Week       int
		TotalPanen int
	}

	// 🔴 VERSI POSTGRES: pakai EXTRACT(ISOYEAR) & EXTRACT(WEEK)
	err = db.Model(&models.FasePanen{}).
		Select(`
            EXTRACT(ISOYEAR FROM tanggal_panen_aktual) AS year,
            EXTRACT(WEEK FROM tanggal_panen_aktual) AS week,
            SUM(jumlah_panen) AS total_panen
        `).
		Where("tanggal_panen_aktual IS NOT NULL").
		Where("tanggal_panen_aktual BETWEEN ? AND ?", startDate, endDate).
		// GROUP BY kolom 1 dan 2 di SELECT (year & week)
		Group("1, 2").
		Order("1, 2").
		Scan(&rows).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Map untuk nyimpen total panen per (year-week)
	dataMap := make(map[string]int)
	for _, r := range rows {
		key := fmt.Sprintf("%04d-%02d", r.Year, r.Week)
		dataMap[key] = r.TotalPanen
	}

	labels := make([]string, 0, 6)
	data := make([]int, 0, 6)

	// Loop 6 minggu (paling lama → terbaru)
	for i := 5; i >= 0; i-- {
		weekTime := now.AddDate(0, 0, -7*i)

		// time.Time.ISOWeek() → (isoyear, week) → ini MATCH dengan EXTRACT(ISOYEAR/WEEK)
		year, week := weekTime.ISOWeek()
		key := fmt.Sprintf("%04d-%02d", year, week)

		totalPanen, ok := dataMap[key]
		if !ok {
			totalPanen = 0 // kalau nggak ada data minggu ini → 0
		}

		label := fmt.Sprintf("Minggu %d", 6-i) // bebas mau diganti
		// label := fmt.Sprintf("%d-W%02d", year, week) // alternatif

		labels = append(labels, label)
		data = append(data, totalPanen)
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"data":   data,
	})
}

