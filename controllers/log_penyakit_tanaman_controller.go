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

// PenyakitTanaman swagger model
// @Description Model untuk data penyakit tanaman
type SwaggerPenyakitTanaman struct {
	ID           uint   `json:"id"`
	NamaPenyakit string `json:"nama_penyakit"`
	Deskripsi    string `json:"deskripsi"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// LogPenyakitTanaman swagger model
// @Description Log penyakit terkait tanaman
type SwaggerLogPenyakitTanaman struct {
	ID              uint   `json:"id"`
	Kondisi         string `json:"kondisi"`
	Catatan         string `json:"catatan"`
	Foto            string `json:"foto,omitempty"`
	FotoLogPenyakitID string `json:"foto_log_penyakit_id,omitempty"`
	SaranPerawatan  string `json:"saran_perawatan"`
	TanamanID       uint   `json:"tanaman_id"`
	PenyakitID      uint   `json:"penyakit_id"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// GetAllLogPenyakit godoc
// @Summary Get all log penyakit tanaman
// @Description Mendapatkan daftar semua log penyakit tanaman dengan pagination
// @Tags LogPenyakitTanaman
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=[]SwaggerLogPenyakitTanaman,meta=utils.Pagination} "Berhasil mengambil data"
// @Failure 500 {object} utils.Response "Gagal koneksi atau query database"
// @Router /Log-Penyakit-Tanaman [get]
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
		Order("created_at DESC").
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

// GetLogPenyakitById godoc
// @Summary Get log penyakit tanaman by ID
// @Description Mendapatkan detail log penyakit tanaman berdasarkan ID
// @Tags LogPenyakitTanaman
// @Accept json
// @Produce json
// @Param id path int true "ID Log Penyakit Tanaman"
// @Success 200 {object} utils.Response{data=SwaggerLogPenyakitTanaman} "Detail ditemukan"
// @Failure 404 {object} utils.Response "Data tidak ditemukan"
// @Failure 500 {object} utils.Response "Gagal retrieve data"
// @Router /Log-Penyakit-Tanaman/{id} [get]
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

// GetLogPenyakitByTanamanId godoc
// @Summary Get log penyakit tanaman by Tanaman ID
// @Description Mendapatkan semua log penyakit berdasarkan tanaman_id dengan pagination
// @Tags LogPenyakitTanaman
// @Accept json
// @Produce json
// @Param id_tanaman path int true "ID Tanaman"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=[]SwaggerLogPenyakitTanaman,meta=utils.Pagination} "Berhasil mengambil data"
// @Failure 404 {object} utils.Response "Data tidak ditemukan"
// @Failure 500 {object} utils.Response "Gagal koneksi atau query database"
// @Router /Log-Penyakit-Tanaman/Tanaman/{id_tanaman} [get]
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