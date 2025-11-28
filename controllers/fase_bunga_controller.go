package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
)

// Helper function to validate MingguKe (must be positive integer)
func isValidMingguKe(mingguKe int) bool {
	return mingguKe > 0
}

// Helper function to parse and validate TanggalCatat (YYYY-MM-DD, not in future)
func parseAndValidateTanggalCatat(dateStr string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("format tanggal_catat harus YYYY-MM-DD")
	}
	if parsed.After(time.Now()) {
		return time.Time{}, fmt.Errorf("tanggal_catat tidak boleh di masa depan")
	}
	return parsed, nil
}

// Helper function to ensure Tanaman exists
func ensureTanamanExists(db *gorm.DB, tanamanID uint) *string {
	if tanamanID == 0 {
		msg := "tanaman_id wajib dan harus > 0"
		return &msg
	}
	var tanaman models.Tanaman
	if err := db.First(&tanaman, tanamanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			msg := "tanaman_id tidak ditemukan"
			return &msg
		}
		msg := "gagal cek tanaman"
		return &msg
	}
	return nil
}

// GET /fase-bunga (with pagination)
func GetAllFaseBunga(c *gin.Context) {
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var totalRows int64
	if err := db.Model(&models.FaseBunga{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hitung data fase bunga", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, perPage, totalRows)
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		utils.ErrorResponseWithData(c, http.StatusBadRequest,
			fmt.Sprintf("Halaman %d di luar jangkauan. Hanya %d halaman tersedia", page, pagination.TotalPages),
			nil,
			"Halaman di luar jangkauan",
		)
		return
	}

	var faseBungaList []models.FaseBunga
	if err := db.Preload("Tanaman").
		Limit(perPage).
		Offset(offset).
		Find(&faseBungaList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data fase bunga", err.Error())
		return
	}

	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data fase bunga", []models.FaseBunga{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Data fase bunga berhasil diambil", faseBungaList, pagination)
}

// GET /fase-bunga/:id
func GetFaseBungaByID(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var faseBunga models.FaseBunga
	if err := db.Preload("Tanaman").First(&faseBunga, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Fase bunga tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase bunga", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail fase bunga", faseBunga)
}

// POST /fase-bunga
func CreateFaseBunga(c *gin.Context) {
	var input struct {
		MingguKe     int    `json:"minggu_ke" binding:"required"`
		TanggalCatat string `json:"tanggal_catat" binding:"required"` // YYYY-MM-DD
		JumlahBunga  int    `json:"jumlah_bunga"`
		BungaPecah   int    `json:"bunga_pecah"`
		PentilMuncul int    `json:"pentil_muncul"`
		TanamanID    uint   `json:"tanaman_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// Validasi MingguKe
	if !isValidMingguKe(input.MingguKe) {
		utils.ErrorResponse(c, http.StatusBadRequest, "minggu_ke harus positif", input.MingguKe)
		return
	}

	// Validasi TanggalCatat
	parsedTanggal, err := parseAndValidateTanggalCatat(input.TanggalCatat)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_catat tidak valid", err.Error())
		return
	}

	// Validasi jumlah (tidak negatif)
	if input.JumlahBunga < 0 || input.BungaPecah < 0 || input.PentilMuncul < 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "jumlah_bunga, bunga_pecah, dan pentil_muncul tidak boleh negatif", nil)
		return
	}

	// Pastikan Tanaman ada
	if msg := ensureTanamanExists(db, input.TanamanID); msg != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, *msg, input.TanamanID)
		return
	}

	// Buat FaseBunga
	faseBunga := models.FaseBunga{
		MingguKe:     input.MingguKe,
		TanggalCatat: &parsedTanggal,
		JumlahBunga:  input.JumlahBunga,
		BungaPecah:   input.BungaPecah,
		PentilMuncul: input.PentilMuncul,
		TanamanID:    input.TanamanID,
	}

	if err := db.Create(&faseBunga).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal simpan fase bunga", err.Error())
		return
	}

	// Preload Tanaman
	db.Preload("Tanaman").First(&faseBunga, faseBunga.ID)

	utils.SuccessResponse(c, http.StatusCreated, "Fase bunga berhasil dibuat", faseBunga)
}

// PUT /fase-bunga/:id
func UpdateFaseBunga(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var faseBunga models.FaseBunga
	if err := db.Preload("Tanaman").First(&faseBunga, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Fase bunga tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase bunga", err.Error())
		return
	}

	var input struct {
		MingguKe     *int    `json:"minggu_ke"`
		TanggalCatat *string `json:"tanggal_catat"` // YYYY-MM-DD
		JumlahBunga  *int    `json:"jumlah_bunga"`
		BungaPecah   *int    `json:"bunga_pecah"`
		PentilMuncul *int    `json:"pentil_muncul"`
		TanamanID    *uint   `json:"tanaman_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	// Apply updates with validation
	if input.MingguKe != nil {
		if !isValidMingguKe(*input.MingguKe) {
			utils.ErrorResponse(c, http.StatusBadRequest, "minggu_ke harus positif", *input.MingguKe)
			return
		}
		faseBunga.MingguKe = *input.MingguKe
	}

	if input.TanggalCatat != nil {
		parsedTanggal, err := parseAndValidateTanggalCatat(*input.TanggalCatat)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_catat tidak valid", err.Error())
			return
		}
		faseBunga.TanggalCatat = &parsedTanggal
	}

	if input.JumlahBunga != nil {
		if *input.JumlahBunga < 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "jumlah_bunga tidak boleh negatif", nil)
			return
		}
		faseBunga.JumlahBunga = *input.JumlahBunga
	}

	if input.BungaPecah != nil {
		if *input.BungaPecah < 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "bunga_pecah tidak boleh negatif", nil)
			return
		}
		faseBunga.BungaPecah = *input.BungaPecah
	}

	if input.PentilMuncul != nil {
		if *input.PentilMuncul < 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "pentil_muncul tidak boleh negatif", nil)
			return
		}
		faseBunga.PentilMuncul = *input.PentilMuncul
	}

	if input.TanamanID != nil {
		if msg := ensureTanamanExists(db, *input.TanamanID); msg != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, *msg, *input.TanamanID)
			return
		}
		faseBunga.TanamanID = *input.TanamanID
	}

	if err := db.Save(&faseBunga).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update fase bunga", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fase bunga berhasil diperbarui", faseBunga)
}

// DELETE /fase-bunga/:id
func DeleteFaseBunga(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var faseBunga models.FaseBunga
	if err := db.First(&faseBunga, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Fase bunga tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase bunga", err.Error())
		return
	}

	if err := db.Delete(&faseBunga).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus fase bunga", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fase bunga berhasil dihapus", utils.EmptyObj{})
}

// GET /fase-bunga/tanaman/:tanaman_id (paginated)
func GetFaseBungaByTanaman(c *gin.Context) {
	tanamanIDStr := c.Param("tanaman_id")
	tanamanID, err := strconv.ParseUint(tanamanIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "tanaman_id tidak valid", tanamanIDStr)
		return
	}

	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// Pastikan Tanaman ada
	if msg := ensureTanamanExists(db, uint(tanamanID)); msg != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, *msg, tanamanID)
		return
	}

	var totalRows int64
	if err := db.Model(&models.FaseBunga{}).Where("tanaman_id = ?", tanamanID).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hitung data fase bunga", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, perPage, totalRows)
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		utils.ErrorResponseWithData(c, http.StatusBadRequest,
			fmt.Sprintf("Halaman %d di luar jangkauan. Hanya %d halaman tersedia", page, pagination.TotalPages),
			nil,
			"Halaman di luar jangkauan",
		)
		return
	}

	var faseBungaList []models.FaseBunga
	if err := db.Where("tanaman_id = ?", tanamanID).
		Preload("Tanaman").
		Limit(perPage).
		Offset(offset).
		Find(&faseBungaList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data fase bunga", err.Error())
		return
	}

	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data fase bunga untuk tanaman ini", []models.FaseBunga{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Data fase bunga berhasil diambil", faseBungaList, pagination)
}
