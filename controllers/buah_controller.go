package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAllBuah godoc
// @Summary Get all buah with pagination
// @Description Retrieve paginated list of buah
// @Tags Buah
// @Security Bearer
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/buah [get]
func GetAllBuah(c *gin.Context){
	// get pagination parameters
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	// count total rows
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to Database", err.Error())
		return
	}

	// count total rows
	var totalRows int64
	if err := db.Model(&models.Buah{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count buah data", err.Error())
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
	var buahList []models.Buah
	if err := db.Preload("Tanaman.Kebun").
		Limit(perPage).
		Offset(offset).
		Find(&buahList).Error; err != nil {
		
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve buah data", err.Error())
			return
	}

	// handle empty data
	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "No buah data found", []models.Buah{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Buah data retrieved successfully", buahList, pagination)

}

// GetBuahByID godoc
// @Summary Get buah by ID
// @Tags Buah
// @Security Bearer
// @Produce json
// @Param id path int true "Buah ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/buah/{id} [get]
func GetBuahByID(c *gin.Context) {
    id := c.Param("id")

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to database", err.Error())
        return
    }

    var buah models.Buah
    if err := db.Preload("Tanaman.Kebun").First(&buah, id).Error; err != nil {
        utils.ErrorResponse(c, http.StatusNotFound, "Buah not found", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "Buah retrieved successfully", buah)
}

// CreateBuah godoc
// @Summary Create a new buah
// @Tags Buah
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body object{nama_buah=string,tanaman_id=int} true "Buah Data"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/buah [post]
func CreateBuah(c *gin.Context) {
    var requestBody struct {
        NamaBuah  string `json:"nama_buah" binding:"required"`
        TanamanID uint   `json:"tanaman_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&requestBody); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err.Error())
        return
    }

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to database", err.Error())
        return
    }

    // Check if tanaman exists
    var tanaman models.Tanaman
    if err := db.First(&tanaman, requestBody.TanamanID).Error; err != nil {
        utils.ErrorResponse(c, http.StatusNotFound, "Tanaman not found", err.Error())
        return
    }

    // Create buah
    buah := models.Buah{
        NamaBuah:  requestBody.NamaBuah,
        TanamanID: requestBody.TanamanID,
    }

    if err := db.Create(&buah).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create buah", err.Error())
        return
    }

    // Preload tanaman data
    db.Preload("Tanaman").First(&buah, buah.ID)

    utils.SuccessResponse(c, http.StatusCreated, "Buah created successfully", buah)
}

// UpdateBuah godoc
// @Summary Update buah
// @Tags Buah
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Buah ID"
// @Param request body object{nama_buah=string,tanaman_id=integer} true "Updated Buah Data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/buah/{id} [put]
func UpdateBuah(c *gin.Context) {
    id := c.Param("id")

    var requestBody struct {
        NamaBuah  string `json:"nama_buah"`
        TanamanID uint   `json:"tanaman_id"`
    }

    if err := c.ShouldBindJSON(&requestBody); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err.Error())
        return
    }

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to database", err.Error())
        return
    }

    // Check if buah exists
    var buah models.Buah
    if err := db.First(&buah, id).Error; err != nil {
        utils.ErrorResponse(c, http.StatusNotFound, "Buah not found", err.Error())
        return
    }

    // Update fields
    if requestBody.NamaBuah != "" {
        buah.NamaBuah = requestBody.NamaBuah
    }
    if requestBody.TanamanID != 0 {
        // Check if tanaman exists
        var tanaman models.Tanaman
        if err := db.First(&tanaman, requestBody.TanamanID).Error; err != nil {
            utils.ErrorResponse(c, http.StatusNotFound, "Tanaman not found", err.Error())
            return
        }
        buah.TanamanID = requestBody.TanamanID
    }

    // Save updates
    if err := db.Save(&buah).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update buah", err.Error())
        return
    }

    // Preload tanaman data
    db.Preload("Tanaman.Kebun").First(&buah, buah.ID)

    utils.SuccessResponse(c, http.StatusOK, "Buah updated successfully", buah)
}

// DeleteBuah godoc
// @Summary Delete buah
// @Tags Buah
// @Security Bearer
// @Produce json
// @Param id path int true "Buah ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/buah/{id} [delete]
func DeleteBuah(c *gin.Context) {
    id := c.Param("id")

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to database", err.Error())
        return
    }

    // Check if buah exists
    var buah models.Buah
    if err := db.First(&buah, id).Error; err != nil {
        utils.ErrorResponse(c, http.StatusNotFound, "Buah not found", err.Error())
        return
    }

    // Delete buah
    if err := db.Delete(&buah).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete buah", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "Buah deleted successfully", utils.EmptyObj{})
}

// GetBuahByKebun godoc
// @Summary Get buah by kebun ID with pagination
// @Description Retrieve buah based on related tanaman kebun
// @Tags Buah
// @Security Bearer
// @Produce json
// @Param id_kebun path int true "Kebun ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/buah/kebun/{id_kebun} [get]
func GetBuahByKebun(c *gin.Context) {
    idKebun := c.Param("id_kebun")

    // Mendapatkan parameter paginasi
    page, perPage := utils.GetPagination(c)
    offset := utils.GetOffset(page, perPage)

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to database", err.Error())
        return
    }

    // 1. Cari semua ID Tanaman yang terkait dengan Kebun ID ini
    var tanamanIDs []uint
    // Menggunakan Pluck untuk mendapatkan daftar ID tanaman
    if err := db.Model(&models.Tanaman{}).
        Where("kebun_id = ?", idKebun).
        Pluck("id", &tanamanIDs).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve related Tanaman IDs", err.Error())
        return
    }

    // 2. Hitung total Buah berdasarkan ID Tanaman yang ditemukan
    var totalRows int64
    if len(tanamanIDs) == 0 {
        totalRows = 0 // Jika tidak ada tanaman, maka tidak ada buah
    } else {
        if err := db.Model(&models.Buah{}).
            // Filter Buah di mana tanaman_id ada di daftar tanamanIDs
            Where("tanaman_id IN (?)", tanamanIDs).
            Count(&totalRows).Error; err != nil {
            utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count buah data", err.Error())
            return
        }
    }
    
    // Perhitungan paginasi
    pagination := utils.CalculatePagination(page, perPage, totalRows)
    
    // Validasi rentang halaman
    if page > pagination.TotalPages && pagination.TotalPages > 0 {
        utils.ErrorResponseWithData(c, http.StatusBadRequest,
            fmt.Sprintf("Page %d out of range. Only %d pages are available", page, pagination.TotalPages),
            nil,
            "Page out of range",
        )
        return
    }

    // 3. Ambil data Buah yang dipaginasi
    var buahList []models.Buah
    if totalRows > 0 {
        if err := db.Preload("Tanaman.Kebun").
            // Filter lagi menggunakan ID Tanaman
            Where("tanaman_id IN (?)", tanamanIDs).
            Limit(perPage).
            Offset(offset).
            Find(&buahList).Error ;err != nil {
            utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve buah data", err.Error())
            return
        }
    }
    
    // Penanganan data kosong
    if totalRows == 0 {
        utils.SuccessResponseWithMeta(c, http.StatusOK, "No buah data found for this kebun", []models.Buah{}, pagination)
        return
    }

    utils.SuccessResponseWithMeta(c, http.StatusOK, "Buah data retrieved successfully", buahList, pagination)
}