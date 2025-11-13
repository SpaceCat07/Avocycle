package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
    db.Preload("Tanaman.Kebun").First(&buah, buah.ID)

    utils.SuccessResponse(c, http.StatusCreated, "Buah created successfully", buah)
}

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

func GetBuahByKebun(c *gin.Context) {
    idKebun := c.Param("id_kebun")

    page, perPage := utils.GetPagination(c)
    offset := utils.GetOffset(page, perPage)

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to database", err.Error())
        return
    }

    var totalRows int64
    if err := db.Model(&models.Buah{}).
        Joins("JOIN tanamen ON tanamen.id = buahs.tanaman_id").
        Where("tanamen.kebun_id = ?", idKebun).
        Count(&totalRows).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count buah data", err.Error())
        return
    }

    pagination := utils.CalculatePagination(page, perPage, totalRows)
    if page > pagination.TotalPages && pagination.TotalPages > 0 {
        utils.ErrorResponseWithData(
            c,
            http.StatusBadRequest,
            fmt.Sprintf("Page %d out of range. Only %d pages are available", page, pagination.TotalPages),
            nil,
            "Page out of range",
        )
        return
    }

    var buahList []models.Buah
    if err := db.Model(&models.Buah{}).
        Joins("JOIN tanamen ON tanamen.id = buahs.tanaman_id").
        Where("tanamen.kebun_id = ?", idKebun).
        Preload("Tanaman.Kebun").
        Offset(offset).
        Limit(perPage).
        Find(&buahList).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve buah data", err.Error())
        return
    }

    utils.SuccessResponseWithMeta(c, http.StatusOK, "Buah data retrieved successfully", buahList, pagination)
}