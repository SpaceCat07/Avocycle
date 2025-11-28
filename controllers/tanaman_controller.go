package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
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
// @Summary Ambil semua tanaman
// @Description Mengambil daftar tanaman dengan pagination
// @Tags Tanaman
// @Produce json
// @Param page query int false "Halaman"
// @Param per_page query int false "Jumlah data per halaman"
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /tanaman [get]
func GetAllTanaman(c *gin.Context) {
	// get pagination parameters
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	// connect to db
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// count total rows
	var totalRows int64
	if err := db.Model(&models.Tanaman{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count tanaman data", err.Error())
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
	var tanamanList []models.Tanaman
	if err := db.Preload("Kebun").
		Limit(perPage).
		Offset(offset).
		Find(&tanamanList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve buah data", err.Error())
		return
	}

	// handle empty data
	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "No tanaman data found", []models.Tanaman{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Tanaman data retrieve successfully", tanamanList, pagination)
}

// GET /tanaman/:id
// @Summary Detail tanaman
// @Description Mendapatkan detail tanaman berdasarkan ID
// @Tags Tanaman
// @Produce json
// @Param id path int true "ID Tanaman"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /tanaman/{id} [get]
func GetTanamanByID(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

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

// POST /
// @Summary Tambah tanaman
// @Description Menambahkan tanaman baru
// @Tags Tanaman
// @Accept multipart/form-data
// @Produce json
// @Param nama_tanaman formData string true "Nama Tanaman"
// @Param varietas formData string true "Varietas"
// @Param tanggal_tanam formData string true "Tanggal Tanam (YYYY-MM-DD)"
// @Param kebun_id formData int true "ID Kebun"
// @Param kode_blok formData string true "Kode Blok"
// @Param kode_tanaman formData string true "Kode Tanaman"
// @Param masa_produksi formData int true "Masa Produksi"
// @Param foto_tanaman formData file false "Foto Tanaman"
// @Security 	Bearer
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /tanaman [post]
func CreateTanaman(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// input struct lokal (tanpa DTO terpisah)
	var input struct {
		NamaTanaman  string `form:"nama_tanaman" binding:"required"`
		Varietas     string `form:"varietas" binding:"required"`
		TanggalTanam string `form:"tanggal_tanam" binding:"required"` // YYYY-MM-DD
		KebunID      uint   `form:"kebun_id" binding:"required"`
		KodeBlok 	 string `form:"kode_blok" binding:"required"`
		KodeTanaman	 string `form:"kode_tanaman" binding:"required"`
		FotoTanaman  *multipart.FileHeader `form:"foto_tanaman"`
		MasaProduksi int	`form:"masa_produksi" binding:"required"`
	}

	// 1) bind JSON
	if err := c.ShouldBind(&input); err != nil {
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
		NamaTanaman:  input.NamaTanaman,
		Varietas:     input.Varietas,
		TanggalTanam: parsedTanggal,
		KebunID:      input.KebunID,
		KodeBlok: 	  input.KodeBlok,
		KodeTanaman:  input.KodeTanaman,
		MasaProduksi: input.MasaProduksi,
	}

	fileHeader, _ := c.FormFile("foto_tanaman")
	if fileHeader != nil {
		url, publicID, uploadErr := utils.AsyncUploadOptionalImage(fileHeader, "tanaman")
		if uploadErr != nil {
            utils.ErrorResponse(c, http.StatusBadRequest, "Upload foto gagal", uploadErr.Error())
            return
        }
        tanaman.FotoTanaman = url
        tanaman.FotoTanamanID = publicID
	}

	if err := db.Create(&tanaman).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan tanaman", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Tanaman berhasil dibuat", tanaman)
}

// PUT /tanaman/:id
// @Summary Update tanaman
// @Description Mengubah data tanaman (bisa sebagian)
// @Tags Tanaman
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID Tanaman"
// @Param nama_tanaman formData string false "Nama Tanaman"
// @Param varietas formData string false "Varietas"
// @Param tanggal_tanam formData string false "Tanggal Tanam"
// @Param kebun_id formData int false "ID Kebun"
// @Param kode_blok formData string false "Kode Blok"
// @Param kode_tanaman formData string false "Kode Tanaman"
// @Param masa_produksi formData int false "Masa Produksi"
// @Param foto_tanaman formData file false "Foto Tanaman"
// @Security 	Bearer
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /tanaman/{id} [put]
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

    var input struct {
        NamaTanaman  *string `form:"nama_tanaman" json:"nama_tanaman"`
        Varietas     *string `form:"varietas" json:"varietas"`
        TanggalTanam *string `form:"tanggal_tanam" json:"tanggal_tanam"`
        KebunID      *uint   `form:"kebun_id" json:"kebun_id"`
        KodeBlok     *string `form:"kode_blok" json:"kode_blok"`
        KodeTanaman  *string `form:"kode_tanaman" json:"kode_tanaman"`
        MasaProduksi *int    `form:"masa_produksi" json:"masa_produksi"`
    }

    if err := c.ShouldBind(&input); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
        return
    }

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

    if input.KodeBlok != nil {
        if kode := strings.TrimSpace(*input.KodeBlok); kode == "" {
            utils.ErrorResponse(c, http.StatusBadRequest, "kode_blok tidak boleh kosong", nil)
            return
        } else {
            tanaman.KodeBlok = kode
        }
    }

    if input.KodeTanaman != nil {
        if kode := strings.TrimSpace(*input.KodeTanaman); kode == "" {
            utils.ErrorResponse(c, http.StatusBadRequest, "kode_tanaman tidak boleh kosong", nil)
            return
        } else {
            tanaman.KodeTanaman = kode
        }
    }

    if input.MasaProduksi != nil {
        if *input.MasaProduksi <= 0 {
            utils.ErrorResponse(c, http.StatusBadRequest, "masa_produksi harus lebih dari 0", nil)
            return
        }
        tanaman.MasaProduksi = *input.MasaProduksi
    }

    fileHeader, _ := c.FormFile("foto_tanaman")
    if fileHeader != nil {
        newURL, newPublicID, uploadErr := utils.AsyncUploadOptionalImage(fileHeader, "tanaman")
        if uploadErr != nil {
            utils.ErrorResponse(c, http.StatusBadRequest, "Upload foto gagal", uploadErr.Error())
            return
        }

        if tanaman.FotoTanamanID != "" {
            if err := utils.DeleteCloudinaryAsset(tanaman.FotoTanamanID); err != nil {
                utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus foto lama", err.Error())
                return
            }
        }

        tanaman.FotoTanaman = newURL
        tanaman.FotoTanamanID = newPublicID
    }

    if err := db.Save(&tanaman).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update tanaman", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "Tanaman berhasil diperbarui", tanaman)
}

// DELETE /tanaman/:id
// @Summary Hapus tanaman
// @Description Menghapus tanaman berdasarkan ID
// @Tags Tanaman
// @Produce json
// @Param id path int true "ID Tanaman"
// @Security 	Bearer
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /tanaman/{id} [delete]
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

    if tanaman.FotoTanamanID != "" {
        if err := utils.DeleteCloudinaryAsset(tanaman.FotoTanamanID); err != nil {
            utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus foto di Cloudinary", err.Error())
            return
        }
    }

    if err := db.Delete(&tanaman).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus tanaman", err.Error())
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "Tanaman berhasil dihapus", utils.EmptyObj{})
}

// @Summary Tanaman berdasarkan Kebun
// @Description Ambil data tanaman berdasarkan kebun_id
// @Tags Tanaman
// @Produce json
// @Param id_kebun path int true "ID Kebun"
// @Param page query int false "Halaman"
// @Param per_page query int false "Jumlah per halaman"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /tanaman/kebun/{id_kebun} [get]
func GetTanamanByKebunID(c *gin.Context) {
	idKebun := c.Param("id_kebun")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	// pagination setup
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	// hitung total tanaman di kebun ini
	var totalRows int64
	if err := db.Model(&models.Tanaman{}).Joins("Tanaman").Where("tanaman.kebun_id = ?", idKebun).Count(&totalRows).Error; err != nil {
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

	// ambil data tanaman berdasarkan kebun_id
	var tanamanList []models.Tanaman
	if err := db.Model(&models.Tanaman{}).Joins("Kebun").Where("tanaman.kebun_id = ?", idKebun).
        Preload("Tanaman").
        Offset(offset).
        Find(&tanamanList).Error ;err != nil {
        
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrive tanaman data", err.Error())
		return
    }

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Tanaman data retrieved successfully", tanamanList, pagination)
}