package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
)

// --- mutex global untuk menghindari race condition ---
var bookingMutex sync.Mutex

// GET /booking
func GetAllBooking(c *gin.Context) {
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var totalRows int64
	if err := db.Model(&models.Booking{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghitung total data booking", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, perPage, totalRows)
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		utils.ErrorResponseWithData(c, http.StatusBadRequest,
			fmt.Sprintf("Page %d out of range. Only %d pages available", page, pagination.TotalPages),
			nil, "Page out of range")
		return
	}

	var bookingList []models.Booking
	if err := db.Preload("User").Preload("Tanaman.Kebun").
		Limit(perPage).
		Offset(offset).
		Find(&bookingList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data booking", err.Error())
		return
	}

	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data booking ditemukan", []models.Booking{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Data booking berhasil diambil", bookingList, pagination)
}

// GET /booking/:id
func GetBookingByID(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var booking models.Booking
	if err := db.Preload("User").Preload("Tanaman.Kebun").First(&booking, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Booking tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data booking", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail booking berhasil diambil", booking)
}

// POST /booking
func CreateBooking(c *gin.Context) {
	bookingMutex.Lock()         // --- mulai critical section ---
	defer bookingMutex.Unlock() // --- akhiri critical section ---

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var input struct {
		UserID    uint `json:"user_id" binding:"required"`
		TanamanID uint `json:"tanaman_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	// Validasi user
	var user models.User
	if err := db.First(&user, input.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusBadRequest, "User tidak ditemukan", input.UserID)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal cek user", err.Error())
		return
	}

	// Validasi tanaman
	var tanaman models.Tanaman
	if err := db.First(&tanaman, input.TanamanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusBadRequest, "Tanaman tidak ditemukan", input.TanamanID)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal cek tanaman", err.Error())
		return
	}

	// Cegah double booking: tanaman hanya boleh dibooking satu kali
	var existingBooking models.Booking
	if err := db.Where("tanaman_id = ?", input.TanamanID).First(&existingBooking).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tanaman ini sudah dibooking", input.TanamanID)
		return
	}

	newBooking := models.Booking{
		UserID:    input.UserID,
		TanamanID: input.TanamanID,
	}

	if err := db.Create(&newBooking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat booking", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Booking berhasil dibuat", newBooking)
}

// PUT /booking/:id
func UpdateBooking(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var booking models.Booking
	if err := db.First(&booking, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Booking tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data booking", err.Error())
		return
	}

	var input struct {
		UserID    *uint `json:"user_id"`
		TanamanID *uint `json:"tanaman_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	// Gunakan mutex juga untuk update booking agar aman dari race condition
	bookingMutex.Lock()
	defer bookingMutex.Unlock()

	if input.UserID != nil {
		var user models.User
		if err := db.First(&user, *input.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusBadRequest, "User tidak ditemukan", *input.UserID)
				return
			}
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal cek user", err.Error())
			return
		}
		booking.UserID = *input.UserID
	}

	if input.TanamanID != nil {
		var tanaman models.Tanaman
		if err := db.First(&tanaman, *input.TanamanID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusBadRequest, "Tanaman tidak ditemukan", *input.TanamanID)
				return
			}
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal cek tanaman", err.Error())
			return
		}

		// Cegah double booking
		var existing models.Booking
		if err := db.Where("tanaman_id = ? AND id != ?", *input.TanamanID, id).First(&existing).Error; err == nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Tanaman sudah dibooking oleh pengguna lain", *input.TanamanID)
			return
		}
		booking.TanamanID = *input.TanamanID
	}

	if err := db.Save(&booking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update booking", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Booking berhasil diperbarui", booking)
}

// DELETE /booking/:id
func DeleteBooking(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var booking models.Booking
	if err := db.First(&booking, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Booking tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data booking", err.Error())
		return
	}

	if err := db.Delete(&booking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus booking", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Booking berhasil dihapus", utils.EmptyObj{})
}

// GET /booking/user/:user_id
func GetBookingByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	uid, err := strconv.Atoi(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "user_id tidak valid", userID)
		return
	}

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	var totalRows int64
	if err := db.Model(&models.Booking{}).Where("user_id = ?", uid).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghitung total data booking user", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, perPage, totalRows)
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		utils.ErrorResponseWithData(c, http.StatusBadRequest,
			fmt.Sprintf("Page %d out of range. Only %d pages available", page, pagination.TotalPages),
			nil, "Page out of range")
		return
	}

	var userBookings []models.Booking
	if err := db.Preload("User").Preload("Tanaman.Kebun").
		Where("user_id = ?", uid).
		Limit(perPage).
		Offset(offset).
		Find(&userBookings).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data booking user", err.Error())
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Data booking user berhasil diambil", userBookings, pagination)
}
