package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllLogPenyakit(c *gin.Context) {
	// get pagination parameters
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	// connect to db
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// count total rows
	var totalRows int64
	if err := db.Model(&models.LogPenyakitTanaman{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count log penyakit tanaman data", err.Error())
		return
	}

	// calculate pagination
	pagination := utils.CalculatePagination(page, perPage, totalRows)

	// validate page range
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		utils.ErrorResponseWithData(c, http.StatusBadRequest,
		fmt.Sprintf("Page %d out of range. Only %d pages are available", page, pagination.TotalPages),
		nil,
		"Page out of range",
	)
	return
	}

	// get paginated data
	var logPenyakitList []models.LogPenyakitTanaman
	if err := db.Preload("Tanaman").
		Preload("Penyakit").
		Limit(perPage).
		Offset(offset).
		Find(&logPenyakitList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve log penyakit tanaman data", err.Error())
		return
	}

	// handle empty data
	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "No log penyakit tanaman found", []models.LogPenyakitTanaman{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Tanaman data retrieve successfully", logPenyakitList, pagination)

}

func GetLogPenyakitById(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to DB", err.Error())
		return
	}

	var logPenyakitTanaman models.LogPenyakitTanaman
	if err := db.Preload("Tanaman").Preload("Penyakit").First(&logPenyakitTanaman, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Log penyakit tanaman tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil log penyakit tanaman", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail log penyakit tanaman", logPenyakitTanaman)
}

func GetLogPenyakitByTanamanId(c *gin.Context) {
	idTanaman := c.Param("id_tanaman")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to DB", err.Error())
		return
	}

	// get pagination parameters
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	// count total rows
	var totalRows int64
	if err := db.Model(&models.LogPenyakitTanaman{}).
				Where("tanaman_id = ?", idTanaman).
				Count(&totalRows).
				Error; err != nil {
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count log penyakit tanaman data", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, perPage, totalRows)
    if page > pagination.TotalPages && pagination.TotalPages > 0 {
        utils.ErrorResponseWithData(c, http.StatusBadRequest,
            fmt.Sprintf("Page %d out of range. Only %d pages are available", page, pagination.TotalPages),
            nil,
            "Page out of range",
        )
        return
    }

	var logPenyakitTanamanList []models.LogPenyakitTanaman
	if err := db.Model(&models.LogPenyakitTanaman{}).
				Where("tanaman_id = ?", idTanaman).
				Preload("Tanaman").
				Preload("Penyakit").
				Offset(offset).
				Order("created_at DESC").
				Find(&logPenyakitTanamanList).Error; err != nil {
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrive log penyakit tanaman data", err.Error())
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Log penyakit tanaman retrieved successfully", logPenyakitTanamanList, pagination)
}