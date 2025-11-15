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

// Helper untuk validasi tanggal cover (YYYY-MM-DD). Boleh tanggal masa depan.
func parseAndValidateTanggalCover(dateStr string) (time.Time, error) {
    parsed, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return time.Time{}, fmt.Errorf("format tanggal_cover harus YYYY-MM-DD")
    }
    return parsed, nil
}

// Helper untuk validasi estimasi panen (YYYY-MM-DD, tidak boleh di masa lalu)
func parseAndValidateEstimasiPanen(dateStr string) (time.Time, error) {
    parsed, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return time.Time{}, fmt.Errorf("format estimasi_panen harus YYYY-MM-DD")
    }
    // Normalize to date-only comparison
    today := time.Now().Truncate(24 * time.Hour)
    if parsed.Before(today) {
        return time.Time{}, fmt.Errorf("estimasi_panen tidak boleh di masa lalu")
    }
    return parsed, nil
}

// GET /fase-berbuah (with pagination)
func GetAllFaseBuah(c *gin.Context) {
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var totalRows int64
	if err := db.Model(&models.FaseBuah{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hitung data fase berbuah", err.Error())
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

	var faseBuahList []models.FaseBuah
	if err := db.Preload("Tanaman").
		Limit(perPage).
		Offset(offset).
		Find(&faseBuahList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data fase berbuah", err.Error())
		return
	}

	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data fase berbuah", []models.FaseBuah{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Data fase berbuah berhasil diambil", faseBuahList, pagination)
}

// GET /fase-berbuah/:id
func GetFaseBuahByID(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var faseBuah models.FaseBuah
	if err := db.Preload("Tanaman").First(&faseBuah, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Fase berbuah tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase berbuah", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail fase berbuah", faseBuah)
}

// POST /fase-berbuah
func CreateFaseBuah(c *gin.Context) {
	var input struct {
		MingguKe      int    `json:"minggu_ke" binding:"required"`
		TanggalCatat  string `json:"tanggal_catat" binding:"required"` // YYYY-MM-DD
		TanggalCover  string `json:"tanggal_cover" binding:"required"` // YYYY-MM-DD
		JumlahCover   int    `json:"jumlah_cover" binding:"required"`
		WarnaLabel    string `json:"warna_label,omitempty"`
		EstimasiPanen string `json:"estimasi_panen" binding:"required"` // YYYY-MM-DD
		TanamanID     uint   `json:"tanaman_id" binding:"required"`
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
	parsedTanggalCatat, err := parseAndValidateTanggalCatat(input.TanggalCatat)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_catat tidak valid", err.Error())
		return
	}

	// Validasi TanggalCover
	parsedTanggalCover, err := parseAndValidateTanggalCover(input.TanggalCover)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_cover tidak valid", err.Error())
		return
	}

	// Validasi EstimasiPanen
	parsedEstimasiPanen, err := parseAndValidateEstimasiPanen(input.EstimasiPanen)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "estimasi_panen tidak valid", err.Error())
		return
	}

	// Validasi JumlahCover (tidak negatif)
	if input.JumlahCover < 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "jumlah_cover tidak boleh negatif", nil)
		return
	}

	// Pastikan Tanaman ada
	if msg := ensureTanamanExists(db, input.TanamanID); msg != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, *msg, input.TanamanID)
		return
	}

	// Buat FaseBuah
	faseBuah := models.FaseBuah{
		MingguKe:      input.MingguKe,
		TanggalCatat:  &parsedTanggalCatat,
		TanggalCover:  &parsedTanggalCover,
		JumlahCover:   input.JumlahCover,
		WarnaLabel:    input.WarnaLabel,
		EstimasiPanen: &parsedEstimasiPanen,
		TanamanID:     input.TanamanID,
	}

	if err := db.Create(&faseBuah).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal simpan fase berbuah", err.Error())
		return
	}

	// Preload Tanaman
	db.Preload("Tanaman").First(&faseBuah, faseBuah.ID)

	utils.SuccessResponse(c, http.StatusCreated, "Fase berbuah berhasil dibuat", faseBuah)
}

// PUT /fase-berbuah/:id
func UpdateFaseBuah(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var faseBuah models.FaseBuah
	if err := db.Preload("Tanaman").First(&faseBuah, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Fase berbuah tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase berbuah", err.Error())
		return
	}

	var input struct {
		MingguKe      *int    `json:"minggu_ke"`
		TanggalCatat  *string `json:"tanggal_catat"` // YYYY-MM-DD
		TanggalCover  *string `json:"tanggal_cover"` // YYYY-MM-DD
		JumlahCover   *int    `json:"jumlah_cover"`
		WarnaLabel    *string `json:"warna_label"`
		EstimasiPanen *string `json:"estimasi_panen"` // YYYY-MM-DD
		TanamanID     *uint   `json:"tanaman_id"`
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
		faseBuah.MingguKe = *input.MingguKe
	}

	if input.TanggalCatat != nil {
		parsedTanggal, err := parseAndValidateTanggalCatat(*input.TanggalCatat)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_catat tidak valid", err.Error())
			return
		}
		faseBuah.TanggalCatat = &parsedTanggal
	}

	if input.TanggalCover != nil {
		parsedTanggal, err := parseAndValidateTanggalCover(*input.TanggalCover)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_cover tidak valid", err.Error())
			return
		}
		faseBuah.TanggalCover = &parsedTanggal
	}

	if input.EstimasiPanen != nil {
		parsedTanggal, err := parseAndValidateEstimasiPanen(*input.EstimasiPanen)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "estimasi_panen tidak valid", err.Error())
			return
		}
		faseBuah.EstimasiPanen = &parsedTanggal
	}

	if input.JumlahCover != nil {
		if *input.JumlahCover < 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "jumlah_cover tidak boleh negatif", nil)
			return
		}
		faseBuah.JumlahCover = *input.JumlahCover
	}

	if input.WarnaLabel != nil {
		faseBuah.WarnaLabel = *input.WarnaLabel
	}

	if input.TanamanID != nil {
		if msg := ensureTanamanExists(db, *input.TanamanID); msg != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, *msg, *input.TanamanID)
			return
		}
		faseBuah.TanamanID = *input.TanamanID
	}

	if err := db.Save(&faseBuah).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update fase berbuah", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fase berbuah berhasil diperbarui", faseBuah)
}

// DELETE /fase-berbuah/:id
func DeleteFaseBuah(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var faseBuah models.FaseBuah
	if err := db.First(&faseBuah, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Fase berbuah tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase berbuah", err.Error())
		return
	}

	if err := db.Delete(&faseBuah).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus fase berbuah", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fase berbuah berhasil dihapus", utils.EmptyObj{})
}

// GET /fase-berbuah/tanaman/:tanaman_id (paginated)
func GetFaseBuahByTanaman(c *gin.Context) {
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
	if err := db.Model(&models.FaseBuah{}).Where("tanaman_id = ?", tanamanID).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hitung data fase berbuah", err.Error())
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

	var faseBuahList []models.FaseBuah
	if err := db.Where("tanaman_id = ?", tanamanID).
		Preload("Tanaman").
		Limit(perPage).
		Offset(offset).
		Find(&faseBuahList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data fase berbuah", err.Error())
		return
	}

	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data fase berbuah untuk tanaman ini", []models.FaseBuah{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Data fase berbuah berhasil diambil", faseBuahList, pagination)
}
