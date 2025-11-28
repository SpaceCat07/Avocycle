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

// Helper: parse and validate tanggal panen aktual (YYYY-MM-DD, not future)
func parseAndValidateTanggalPanenAktual(dateStr string) (time.Time, error) {
    parsed, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return time.Time{}, fmt.Errorf("format tanggal_panen_aktual harus YYYY-MM-DD")
    }
    if parsed.After(time.Now()) {
        return time.Time{}, fmt.Errorf("tanggal_panen_aktual tidak boleh di masa depan")
    }
    return parsed, nil
}

// GET /fase-panen (paginated)
func GetAllFasePanen(c *gin.Context) {
    page, perPage := utils.GetPagination(c)
    offset := utils.GetOffset(page, perPage)

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

    var totalRows int64
    if err := db.Model(&models.FasePanen{}).Count(&totalRows).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hitung data fase panen", err.Error())
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

    var list []models.FasePanen
    if err := db.Preload("Tanaman").
        Limit(perPage).
        Offset(offset).
        Find(&list).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data fase panen", err.Error())
        return
    }

    if totalRows == 0 {
        utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data fase panen", []models.FasePanen{}, pagination)
        return
    }

    utils.SuccessResponseWithMeta(c, http.StatusOK, "Data fase panen berhasil diambil", list, pagination)
}

// GET /fase-panen/:id
func GetFasePanenByID(c *gin.Context) {
    id := c.Param("id")

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

    var rec models.FasePanen
    if err := db.Preload("Tanaman").First(&rec, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Fase panen tidak ditemukan", nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase panen", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "Detail fase panen", rec)
}

// POST /fase-panen
func CreateFasePanen(c *gin.Context) {
    var input struct {
        TanggalPanenAktual string  `json:"tanggal_panen_aktual" binding:"required"` // YYYY-MM-DD
        JumlahPanen        int     `json:"jumlah_panen"`
        JumlahSampel       int     `json:"jumlah_sampel"`
        BeratTotal         float64 `json:"berat_total"`
        Catatan            string  `json:"catatan"`
        FotoPanen          string  `json:"foto_panen"`
        TanamanID          uint    `json:"tanaman_id" binding:"required"`
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

    // Validasi tanggal panen aktual
    parsedTanggal, err := parseAndValidateTanggalPanenAktual(input.TanggalPanenAktual)
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_panen_aktual tidak valid", err.Error())
        return
    }

    // Validasi angka tidak negatif
    if input.JumlahPanen < 0 || input.JumlahSampel < 0 || input.BeratTotal < 0 {
        utils.ErrorResponse(c, http.StatusBadRequest, "jumlah_panen, jumlah_sampel, dan berat_total tidak boleh negatif", nil)
        return
    }

    // Pastikan Tanaman ada
    if msg := ensureTanamanExists(db, input.TanamanID); msg != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, *msg, input.TanamanID)
        return
    }

    rec := models.FasePanen{
        TanggalPanenAktual: &parsedTanggal,
        JumlahPanen:        input.JumlahPanen,
        JumlahSampel:       input.JumlahSampel,
        BeratTotal:         input.BeratTotal,
        Catatan:            input.Catatan,
        FotoPanen:          input.FotoPanen,
        TanamanID:          input.TanamanID,
    }

    if err := db.Create(&rec).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal simpan fase panen", err.Error())
        return
    }

    db.Preload("Tanaman").First(&rec, rec.ID)

    utils.SuccessResponse(c, http.StatusCreated, "Fase panen berhasil dibuat", rec)
}

// PUT /fase-panen/:id
func UpdateFasePanen(c *gin.Context) {
    id := c.Param("id")

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

    var rec models.FasePanen
    if err := db.Preload("Tanaman").First(&rec, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Fase panen tidak ditemukan", nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase panen", err.Error())
        return
    }

    var input struct {
        TanggalPanenAktual *string  `json:"tanggal_panen_aktual"` // YYYY-MM-DD
        JumlahPanen        *int     `json:"jumlah_panen"`
        JumlahSampel       *int     `json:"jumlah_sampel"`
        BeratTotal         *float64 `json:"berat_total"`
        Catatan            *string  `json:"catatan"`
        FotoPanen          *string  `json:"foto_panen"`
        TanamanID          *uint    `json:"tanaman_id"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
        return
    }

    if input.TanggalPanenAktual != nil {
        parsed, err := parseAndValidateTanggalPanenAktual(*input.TanggalPanenAktual)
        if err != nil {
            utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_panen_aktual tidak valid", err.Error())
            return
        }
        rec.TanggalPanenAktual = &parsed
    }

    if input.JumlahPanen != nil {
        if *input.JumlahPanen < 0 {
            utils.ErrorResponse(c, http.StatusBadRequest, "jumlah_panen tidak boleh negatif", nil)
            return
        }
        rec.JumlahPanen = *input.JumlahPanen
    }

    if input.JumlahSampel != nil {
        if *input.JumlahSampel < 0 {
            utils.ErrorResponse(c, http.StatusBadRequest, "jumlah_sampel tidak boleh negatif", nil)
            return
        }
        rec.JumlahSampel = *input.JumlahSampel
    }

    if input.BeratTotal != nil {
        if *input.BeratTotal < 0 {
            utils.ErrorResponse(c, http.StatusBadRequest, "berat_total tidak boleh negatif", nil)
            return
        }
        rec.BeratTotal = *input.BeratTotal
    }

    if input.Catatan != nil {
        rec.Catatan = *input.Catatan
    }

    if input.FotoPanen != nil {
        rec.FotoPanen = *input.FotoPanen
    }

    if input.TanamanID != nil {
        if msg := ensureTanamanExists(db, *input.TanamanID); msg != nil {
            utils.ErrorResponse(c, http.StatusBadRequest, *msg, *input.TanamanID)
            return
        }
        rec.TanamanID = *input.TanamanID
    }

    if err := db.Save(&rec).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update fase panen", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "Fase panen berhasil diperbarui", rec)
}

// DELETE /fase-panen/:id
func DeleteFasePanen(c *gin.Context) {
    id := c.Param("id")

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

    var rec models.FasePanen
    if err := db.First(&rec, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Fase panen tidak ditemukan", nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil fase panen", err.Error())
        return
    }

    if err := db.Delete(&rec).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus fase panen", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "Fase panen berhasil dihapus", utils.EmptyObj{})
}

// GET /fase-panen/tanaman/:tanaman_id (paginated)
func GetFasePanenByTanaman(c *gin.Context) {
    tanamanIDStr := c.Param("tanaman_id")
    tanamanIDU64, err := strconv.ParseUint(tanamanIDStr, 10, 32)
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "tanaman_id tidak valid", tanamanIDStr)
        return
    }
    tanamanID := uint(tanamanIDU64)

    page, perPage := utils.GetPagination(c)
    offset := utils.GetOffset(page, perPage)

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

    // Pastikan Tanaman ada
    if msg := ensureTanamanExists(db, tanamanID); msg != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, *msg, tanamanID)
        return
    }

    var totalRows int64
    if err := db.Model(&models.FasePanen{}).Where("tanaman_id = ?", tanamanID).Count(&totalRows).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hitung data fase panen", err.Error())
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

    var list []models.FasePanen
    if err := db.Where("tanaman_id = ?", tanamanID).
        Preload("Tanaman").
        Limit(perPage).
        Offset(offset).
        Find(&list).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data fase panen", err.Error())
        return
    }

    if totalRows == 0 {
        utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data fase panen untuk tanaman ini", []models.FasePanen{}, pagination)
        return
    }

    utils.SuccessResponseWithMeta(c, http.StatusOK, "Data fase panen berhasil diambil", list, pagination)
}
