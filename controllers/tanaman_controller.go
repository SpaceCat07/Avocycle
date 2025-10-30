package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
)

func isValidVarietas(v string) bool {
	switch v {
	case "Var1", "Var2", "Var3":
		return true
	default:
		return false
	}
}

// parse tanggal dan pastikan tidak di masa depan
func parseAndValidateTanggalTanam(dateStr string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("format harus YYYY-MM-DD")
	}
	if parsed.After(time.Now()) {
		return time.Time{}, fmt.Errorf("tanggal_tanam tidak boleh di masa depan")
	}
	return parsed, nil
}

// pastikan kebun dengan ID tertentu ada
func ensureKebunExists(db *gorm.DB, kebunID uint) *string {
	if kebunID == 0 {
		msg := "kebun_id wajib dan harus > 0"
		return &msg
	}
	var kebun models.Kebun
	if err := db.First(&kebun, kebunID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			msg := "kebun_id tidak ditemukan"
			return &msg
		}
		msg := "gagal cek kebun"
		return &msg
	}
	return nil
}

// --- controller ---
func GetAllTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	pageParam := c.DefaultQuery("page", "1")
	limitParam := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := db.Model(&models.Tanaman{}).Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hitung data", err.Error())
		return
	}

	var tanaman []models.Tanaman
	if err := db.Preload("Kebun").Limit(limit).Offset(offset).Find(&tanaman).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data tanaman", err.Error())
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	utils.SuccessResponse(c, http.StatusOK, "Daftar tanaman", gin.H{
		"items": tanaman,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GET /tanaman/:id
func GetTanamanByID(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	id := c.Param("id")
	var tanaman models.Tanaman
	if err := db.Preload("Kebun").First(&tanaman, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Tanaman tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil tanaman", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail tanaman", tanaman)
}

// POST /tanaman
func CreateTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// input struct lokal (tanpa DTO terpisah)
	var input struct {
		NamaTanaman  string `json:"nama_tanaman" binding:"required"`
		Varietas     string `json:"varietas" binding:"required"`
		TanggalTanam string `json:"tanggal_tanam" binding:"required"` // YYYY-MM-DD
		KebunID      uint   `json:"kebun_id" binding:"required"`
	}

	// 1) bind JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	// 2) validasi field
	nama := strings.TrimSpace(input.NamaTanaman)
	if nama == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "nama_tanaman tidak boleh kosong", nil)
		return
	}

	if !isValidVarietas(input.Varietas) {
		utils.ErrorResponse(c, http.StatusBadRequest, "varietas tidak valid. Hanya boleh Var1 / Var2 / Var3", input.Varietas)
		return
	}

	parsedTanggal, err := parseAndValidateTanggalTanam(input.TanggalTanam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_tanam tidak valid", err.Error())
		return
	}

	if msg := ensureKebunExists(db, input.KebunID); msg != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, *msg, input.KebunID)
		return
	}

	// 3) map ke model & simpan
	tanaman := models.Tanaman{
		NamaTanaman:  nama,
		Varietas:     input.Varietas,
		TanggalTanam: parsedTanggal,
		KebunID:      input.KebunID,
	}

    if err := db.Create(&tanaman).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan tanaman", err.Error())
        return
    }

    // preload Kebun for response
    if err := db.Preload("Kebun").First(&tanaman, tanaman.ID).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memuat relasi kebun", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusCreated, "Tanaman berhasil dibuat", tanaman)
}

// PUT /tanaman/:id
func UpdateTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	id := c.Param("id")

    var tanaman models.Tanaman
    if err := db.Preload("Kebun").First(&tanaman, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Tanaman tidak ditemukan", nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil tanaman", err.Error())
        return
    }

	// input struct lokal (pointer untuk partial update)
	var input struct {
		NamaTanaman  *string `json:"nama_tanaman"`
		Varietas     *string `json:"varietas"`
		TanggalTanam *string `json:"tanggal_tanam"` // YYYY-MM-DD
		KebunID      *uint   `json:"kebun_id"`
	}

	// 1) bind JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	// 2) apply+validasi hanya field yang dikirim
	if input.NamaTanaman != nil {
		nama := strings.TrimSpace(*input.NamaTanaman)
		if nama == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "nama_tanaman tidak boleh kosong", nil)
			return
		}
		tanaman.NamaTanaman = nama
	}

	if input.Varietas != nil {
		if !isValidVarietas(*input.Varietas) {
			utils.ErrorResponse(c, http.StatusBadRequest, "varietas tidak valid. Hanya boleh Var1 / Var2 / Var3", *input.Varietas)
			return
		}
		tanaman.Varietas = *input.Varietas
	}

	if input.TanggalTanam != nil {
		if strings.TrimSpace(*input.TanggalTanam) == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_tanam tidak boleh kosong", nil)
			return
		}
		parsedTanggal, err := parseAndValidateTanggalTanam(*input.TanggalTanam)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "tanggal_tanam tidak valid", err.Error())
			return
		}
		tanaman.TanggalTanam = parsedTanggal
	}

	if input.KebunID != nil {
		if msg := ensureKebunExists(db, *input.KebunID); msg != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, *msg, *input.KebunID)
			return
		}
		tanaman.KebunID = *input.KebunID
	}

	// 3) simpan
	if err := db.Save(&tanaman).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update tanaman", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tanaman berhasil diperbarui", tanaman)
}

// DELETE /tanaman/:id
func DeleteTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	id := c.Param("id")

    var tanaman models.Tanaman
    if err := db.Preload("Kebun").First(&tanaman, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Tanaman tidak ditemukan", nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil tanaman", err.Error())
        return
    }

	if err := db.Delete(&tanaman).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus tanaman", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tanaman berhasil dihapus", nil)
}
