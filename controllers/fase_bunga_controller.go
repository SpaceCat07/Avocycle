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

type CreateFaseBungaInput struct {
    MingguKe      int    `json:"minggu_ke" example:"3"`
    TanggalCatat  string `json:"tanggal_catat" example:"2025-12-20"` // format ISO-8601
    JumlahBunga   int    `json:"jumlah_bunga" example:"10"`
    BungaPecah    int    `json:"bunga_pecah" example:"2"`
    PentilMuncul  int    `json:"pentil_muncul" example:"1"`
    TanamanID     uint   `json:"tanaman_id" example:"5"`
}

type UpdateFaseBungaInput struct {
    MingguKe      *int    `json:"minggu_ke" example:"3"`
    TanggalCatat  *string `json:"tanggal_catat" example:"2025-12-20"`
    JumlahBunga   *int    `json:"jumlah_bunga" example:"10"`
    BungaPecah    *int    `json:"bunga_pecah" example:"2"`
    PentilMuncul  *int    `json:"pentil_muncul" example:"1"`
    TanamanID     *uint   `json:"tanaman_id" example:"5"`
}

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

// @Summary Get all fase bunga
// @Description Mendapatkan daftar semua fase bunga (dengan pagination)
// @Tags Fase Bunga
// @Security Bearer
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman"
// @Param per_page query int false "Jumlah data per halaman"
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/fase-bunga [get]
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

// @Summary Get fase bunga by ID
// @Description Mendapatkan detail fase bunga berdasarkan ID
// @Tags Fase Bunga
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID Fase Bunga"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /petani/fase-bunga/{id} [get]
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

// @Summary Create fase bunga
// @Description Menambahkan fase bunga baru untuk tanaman
// @Tags Fase Bunga
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body controllers.CreateFaseBungaInput true "Fase Bunga Data"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /petani/fase-bunga [post]
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

// @Summary Update fase bunga
// @Description Mengupdate data fase bunga berdasarkan ID
// @Tags Fase Bunga
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID Fase Bunga"
// @Param request body controllers.UpdateFaseBungaInput true "Fase Bunga Data"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /petani/fase-bunga/{id} [put]
func UpdateFaseBunga(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// 1. Ambil data eksisting beserta relasinya
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
		MingguKe      *int    `json:"minggu_ke"`
		TanggalCatat  *string `json:"tanggal_catat"` // YYYY-MM-DD
		JumlahBunga   *int    `json:"jumlah_bunga"`
		BungaPecah    *int    `json:"bunga_pecah"`
		PentilMuncul  *int    `json:"pentil_muncul"`
		TanamanID     *uint   `json:"tanaman_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	// Apply updates with validation
	if input.MingguKe != nil {
		// Asumsi isValidMingguKe, models, dan utils sudah didefinisikan
		if !isValidMingguKe(*input.MingguKe) {
			utils.ErrorResponse(c, http.StatusBadRequest, "minggu_ke harus positif", *input.MingguKe)
			return
		}
		faseBunga.MingguKe = *input.MingguKe
	}

	if input.TanggalCatat != nil {
		// Asumsi parseAndValidateTanggalCatat sudah didefinisikan
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

	// 2. Logika Update Tanaman ID (FIX UTAMA)
	if input.TanamanID != nil {
		// Pastikan tanaman baru valid di DB
		if msg := ensureTanamanExists(db, *input.TanamanID); msg != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, *msg, *input.TanamanID)
			return
		}
		
		// Update Foreign Key di struct
		faseBunga.TanamanID = *input.TanamanID

		// PENTING: Kosongkan struct relasi Tanaman yang di-preload.
		// Ini mencegah GORM mengembalikan 'tanaman_id' ke nilai lama saat db.Save().
		faseBunga.Tanaman = models.Tanaman{}
	}

	// 3. Simpan Perubahan
	if err := db.Save(&faseBunga).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update fase bunga", err.Error())
		return
	}

	// 4. Muat Ulang Data (Re-fetch)
	// Muat ulang dengan Preload agar struct Tanaman terisi dengan data TanamanID yang BARU
	if err := db.
		Model(&faseBunga).
		Select("fase_bungas.*").
		Preload("Tanaman").
		Preload("Tanaman.Kebun"). // Opsional: jika relasi ke kebun juga dibutuhkan
		First(&faseBunga, faseBunga.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal muat ulang data", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fase bunga berhasil diperbarui", faseBunga)
}

// @Summary Delete fase bunga
// @Description Menghapus fase bunga berdasarkan ID
// @Tags Fase Bunga
// @Security Bearer
// @Param id path int true "ID Fase Bunga"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /petani/fase-bunga/{id} [delete]
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

// @Summary Get fase bunga by tanaman ID
// @Description Mendapatkan semua fase bunga milik tanaman tertentu
// @Tags Fase Bunga
// @Security Bearer
// @Param tanaman_id path int true "ID Tanaman"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /petani/fase-bunga/tanaman/{tanaman_id} [get]
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
