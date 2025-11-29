package controllers

import (
    "fmt"
    "net/http"
    "strconv"
    "time"
    "mime/multipart"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "Avocycle/config"
    "Avocycle/models"
    "Avocycle/utils"
)

type CreateFasePanenInput struct {
    TanggalPanenAktual string  `json:"tanggal_panen_aktual" example:"2025-11-28"`
    JumlahPanen        int     `json:"jumlah_panen" example:"50"`
    JumlahSampel       int     `json:"jumlah_sampel" example:"5"`
    BeratTotal         float64 `json:"berat_total" example:"22.5"`
    Catatan            string  `json:"catatan" example:"Hasil panen bagus"`
    FotoPanen          string  `json:"foto_panen" example:"https://example.com/img.jpg"`
    FotoPanenID        string  `json:"foto_panen_id" example:"IMG1234"`
    TanamanID          uint    `json:"tanaman_id" example:"3"`
}

type UpdateFasePanenInput struct {
    TanggalPanenAktual *string  `json:"tanggal_panen_aktual" example:"2025-11-29"`
    JumlahPanen        *int     `json:"jumlah_panen" example:"60"`
    JumlahSampel       *int     `json:"jumlah_sampel" example:"7"`
    BeratTotal         *float64 `json:"berat_total" example:"25.1"`
    Catatan            *string  `json:"catatan" example:"Kualitas sedang"`
    FotoPanen          *string  `json:"foto_panen" example:"https://example.com/new.jpg"`
    FotoPanenID        *string  `json:"foto_panen_id" example:"IMG1235"`
    TanamanID          *uint    `json:"tanaman_id" example:"3"`
}

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
// @Summary Get all fase panen
// @Description Mendapatkan semua fase panen dengan pagination
// @Tags Fase Panen
// @Security Bearer
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman"
// @Param per_page query int false "Jumlah data per halaman"
// @Success 200 {object} utils.Response
// @Router /petani/fase-panen [get]
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
// @Summary Get fase panen by ID
// @Description Mendapatkan detail fase panen
// @Tags Fase Panen
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID Fase Panen"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /petani/fase-panen/{id} [get]
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
// @Summary Create fase panen
// @Description Menambahkan data fase panen baru
// @Tags Fase Panen
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param tanggal_panen_aktual formData string true "Tanggal Panen Aktual (YYYY-MM-DD)"
// @Param jumlah_panen formData int false "Jumlah Panen"
// @Param jumlah_sampel formData int false "Jumlah Sampel"
// @Param berat_total formData number false "Berat Total (Kg)"
// @Param catatan formData string false "Catatan"
// @Param foto_panen formData file false "Foto Panen"
// @Param tanaman_id formData int true "ID Tanaman"
// @Success 201 {object} utils.Response
// @Router /petani/fase-panen [post]
func CreateFasePanen(c *gin.Context) {
    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

    var input struct {
        TanggalPanenAktual string `form:"tanggal_panen_aktual" binding:"required"`
        JumlahPanen        int    `form:"jumlah_panen"`
        JumlahSampel       int    `form:"jumlah_sampel"`
        BeratTotal         float64`form:"berat_total"`
        Catatan            string `form:"catatan"`
        TanamanID          uint   `form:"tanaman_id" binding:"required"`
        FotoPanen          *multipart.FileHeader `form:"foto_panen"`
    }

    if err := c.ShouldBind(&input); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
        return
    }

    parsedTanggal, err := parseAndValidateTanggalPanenAktual(input.TanggalPanenAktual)
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_panen_aktual tidak valid", err.Error())
        return
    }

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
        TanamanID:          input.TanamanID,
    }

    // Upload foto jika ada
    if input.FotoPanen != nil {
        url, publicID, errUpload := utils.AsyncUploadOptionalImage(input.FotoPanen, "panen")
        if errUpload != nil {
            utils.ErrorResponse(c, http.StatusBadRequest, "Upload foto gagal", errUpload.Error())
            return
        }
        rec.FotoPanen = url
        rec.FotoPanenID = publicID
    }

    if err := db.Create(&rec).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal simpan data fase panen", err.Error())
		return
	}

    var createdRec models.FasePanen
	if err := db.Preload("Tanaman").First(&createdRec, rec.ID).Error; err != nil {
		// Jika gagal mengambil ulang (walaupun jarang terjadi), kembalikan respons error
		// atau cukup kembalikan rec yang asli tanpa preload.
		utils.ErrorResponse(c, http.StatusInternalServerError, "Fase panen berhasil dibuat, namun gagal memuat detail Tanaman", err.Error())
		return
	}

    utils.SuccessResponse(c, http.StatusCreated, "Fase panen berhasil dibuat", createdRec)
}

// PUT /fase-panen/:id
// @Summary Update fase panen
// @Description Mengupdate data fase panen
// @Tags Fase Panen
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID Fase Panen"
// @Param tanggal_panen_aktual formData string false "YYYY-MM-DD"
// @Param jumlah_panen formData int false "Jumlah Panen"
// @Param jumlah_sampel formData int false "Jumlah Sampel"
// @Param berat_total formData number false "Berat Total (Kg)"
// @Param catatan formData string false "Catatan"
// @Param foto_panen formData file false "Foto Panen"
// @Param tanaman_id formData int false "ID Tanaman"
// @Success 200 {object} utils.Response
// @Router /petani/fase-panen/{id} [put]
func UpdateFasePanen(c *gin.Context) {
    id := c.Param("id")

    db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

    var rec models.FasePanen
    if err := db.First(&rec, id).Error; err != nil {
        utils.ErrorResponse(c, http.StatusNotFound, "Data tidak ditemukan", nil)
        return
    }

    var input struct {
        TanggalPanenAktual *string `form:"tanggal_panen_aktual"`
        JumlahPanen        *int    `form:"jumlah_panen"`
        JumlahSampel       *int    `form:"jumlah_sampel"`
        BeratTotal         *float64`form:"berat_total"`
        Catatan            *string `form:"catatan"`
        TanamanID          *uint   `form:"tanaman_id"`
        FotoPanen          *multipart.FileHeader `form:"foto_panen"`
    }

    if err := c.ShouldBind(&input); err != nil {
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
        rec.JumlahPanen = *input.JumlahPanen
    }

    if input.JumlahSampel != nil {
        rec.JumlahSampel = *input.JumlahSampel
    }

    if input.BeratTotal != nil {
        rec.BeratTotal = *input.BeratTotal
    }

    if input.Catatan != nil {
        rec.Catatan = *input.Catatan
    }

    if input.TanamanID != nil {
        if msg := ensureTanamanExists(db, *input.TanamanID); msg != nil {
            utils.ErrorResponse(c, http.StatusBadRequest, *msg, *input.TanamanID)
            return
        }
        rec.TanamanID = *input.TanamanID
    }

    // Ganti foto jika ada upload baru
    if input.FotoPanen != nil {
        url, publicID, errUpload := utils.AsyncUploadOptionalImage(input.FotoPanen, "panen")
        if errUpload != nil {
            utils.ErrorResponse(c, http.StatusBadRequest, "Upload foto gagal", errUpload.Error())
            return
        }
        rec.FotoPanen = url
        rec.FotoPanenID = publicID
    }

    if err := db.Save(&rec).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update", err.Error())
		return
	}

    var updatedRec models.FasePanen
	// Ambil data berdasarkan ID dari rec yang baru disimpan
	if err := db.Preload("Tanaman").First(&updatedRec, rec.ID).Error; err != nil {
		// Jika gagal memuat detail Tanaman, kembalikan record rec yang asli 
		// (walaupun tanpa detail Tanaman) agar proses update tetap dianggap berhasil.
		utils.ErrorResponse(c, http.StatusOK, "Fase panen berhasil diperbarui, namun gagal memuat detail Tanaman.", rec)
		return
	}

    utils.SuccessResponse(c, http.StatusOK, "Fase panen diperbarui", updatedRec)
}

// DELETE /fase-panen/:id
// @Summary Delete fase panen
// @Description Menghapus fase panen berdasarkan ID
// @Tags Fase Panen
// @Security Bearer
// @Param id path int true "ID Fase Panen"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /petani/fase-panen/{id} [delete]
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
// @Summary Get fase panen by tanaman ID
// @Description Mendapatkan daftar fase panen milik tanaman tertentu
// @Tags Fase Panen
// @Security Bearer
// @Accept json
// @Produce json
// @Param tanaman_id path int true "ID Tanaman"
// @Param page query int false "Nomor halaman"
// @Param per_page query int false "Jumlah data per halaman"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /petani/fase-panen/tanaman/{tanaman_id} [get]
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
