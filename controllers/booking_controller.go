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

// BookingRequest digunakan hanya untuk Swagger (body POST)
type BookingRequest struct {
    UserID    uint `json:"user_id" example:"1"`
    TanamanID uint `json:"tanaman_id" example:"10"`
}

// --- mutex global untuk menghindari race condition ---
var bookingMutex sync.Mutex

// GetAllBooking godoc
// @Summary Get all booking with pagination
// @Description Retrieve paginated list of booking (hanya untuk role Pembeli)
// @Tags Booking
// @Security Bearer
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} utils.Response{data=[]models.SwaggerBooking,meta=utils.Pagination}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /pembeli/booking [get]
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

// GetBookingByID godoc
// @Summary Get booking by ID
// @Description Retrieve detail booking by ID (role Pembeli)
// @Tags Booking
// @Security Bearer
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} utils.Response{data=models.SwaggerBooking}
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /pembeli/booking/{id} [get]
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

// CreateBooking godoc
// @Summary Create a new booking
// @Description Booking action by pembeli
// @Tags Booking
// @Security Bearer
// @Accept json
// @Produce json
// @Param booking body controllers.BookingRequest true "Booking input"
// @Success 201 {object} utils.Response{data=models.SwaggerBooking}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /pembeli/booking [post]
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

	// Cek role user harus Pembeli
	if user.Role != "Pembeli" {
		utils.ErrorResponse(c, http.StatusForbidden, "Hanya user dengan role Pembeli yang bisa membuat booking", input.UserID)
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

	// Ambil ulang booking lengkap dengan relasi
	var result models.Booking
	if err := db.Preload("User").Preload("Tanaman.Kebun").
		First(&result, newBooking.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memuat detail booking", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Booking berhasil dibuat", result)
}

// UpdateBooking godoc
// @Summary Update existing booking
// @Description Update booking by ID (role Pembeli)
// @Tags Booking
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param booking body controllers.BookingRequest true "Booking input"
// @Success 200 {object} utils.Response{data=models.SwaggerBooking}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /pembeli/booking/{id} [put]
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

	// Reload booking terbaru beserta relasi
	if err := db.Preload("User").Preload("Tanaman").First(&booking, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data booking terbaru", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Booking berhasil diperbarui", booking)
}

// DeleteBooking godoc
// @Summary Delete booking by ID
// @Description Delete action for booking (role Pembeli)
// @Tags Booking
// @Security Bearer
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /pembeli/booking/{id} [delete]
func DeleteBooking(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var booking models.Booking
	if err := db.Preload("User").Preload("Tanaman").First(&booking, id).Error; err != nil {
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

// GetBookingByUserID godoc
// @Summary Get booking list by user ID
// @Description Retrieve user's booking list with pagination
// @Tags Booking
// @Security Bearer
// @Produce json
// @Param user_id path int true "User ID"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} utils.Response{data=[]models.SwaggerBooking,meta=utils.Pagination}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /pembeli/booking/user/{user_id} [get]
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
