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

func GetAllPenyakit(c *gin.Context) {
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
	if err := db.Model(&models.PenyakitTanaman{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count penyakit tanaman data", err.Error())
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
	
	var penyakitList []models.PenyakitTanaman
	if err := db.Limit(perPage).
		Offset(offset).
		Order("created_at DESC").
		Find(&penyakitList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve penyakit tanaman data", err.Error())
		return
	}

	// handle empty data
	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "No penyakit tanaman found", []models.LogPenyakitTanaman{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Penyakit data retrieve successfully", penyakitList, pagination)
}

func GetPenyakitById(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to DB", err.Error())
		return
	}

	var penyakitTanaman models.PenyakitTanaman
	if err := db.First(&penyakitTanaman, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Penyakit tanaman tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil penyakit tanaman", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail penyakit tanaman", penyakitTanaman)
}

func UpdatePenyakitTanaman(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to DB", err.Error())
		return
	}

	var penyakitTanaman models.PenyakitTanaman
	if err := db.First(&penyakitTanaman, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Tanaman tidak ditemukan", nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil tanaman", err.Error())
        return
	}


	var input struct {
		Deskripsi string `form:"deskripsi" json:"deskripsi"`
	}

	if err := c.ShouldBind(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
        return
	}

	if err := db.Save(&penyakitTanaman).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update tanaman", err.Error())
        return
	}

	utils.SuccessResponse(c, http.StatusOK, "penyakit tanaman berhasil diperbarui", penyakitTanaman)
}

func DeletePenyakitTanaman(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to DB", err.Error())
		return
	}

	var penyakitTanaman models.PenyakitTanaman
	if err := db.First(&penyakitTanaman, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Penyakit tanaman tidak ditemukan", nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil penyakit tanaman", err.Error())
        return
	}

	var usage int64
    if err := db.Model(&models.LogPenyakitTanaman{}).
        Where("penyakit_id = ?", penyakitTanaman.ID).
        Count(&usage).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengecek relasi log penyakit", err.Error())
        return
    }

    if usage > 0 {
        utils.ErrorResponse(c, http.StatusConflict,
            "Penyakit masih digunakan pada log penyakit tanaman", gin.H{"referenced_rows": usage})
        return
    }

	if err := db.Delete(&penyakitTanaman).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus penyakit tanaman", err.Error())
        return
    }

	utils.SuccessResponse(c, http.StatusOK, "Tanaman berhasil dihapus", utils.EmptyObj{})
}