package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
)

type UpdateKebunRequest struct {
	NamaKebun *string `json:"nama_kebun"`
	MDPL      *string `json:"mdpl"`
}

// --- CONTROLLERS ---

// GET /kebun
// GetAllKebun godoc
// @Summary      Ambil semua kebun
// @Description  Mengambil daftar kebun dengan pagination (menggunakan meta pagination sesuai utils Pagination)
// @Tags         Kebun
// @Param        page      query   int     false  "Halaman"
// @Param        per_page  query   int     false  "Jumlah data per halaman"
// @Success      200  {object}  utils.Response  "Data kebun berhasil diambil"
// @Failure      500  {object}  utils.Response
// @Router       /kebun [get]
func GetAllKebun(c *gin.Context) {
	page, perPage := utils.GetPagination(c)
	offset := utils.GetOffset(page, perPage)

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var totalRows int64
	if err := db.Model(&models.Kebun{}).Count(&totalRows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghitung data kebun", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, perPage, totalRows)
	if page > pagination.TotalPages && pagination.TotalPages > 0 {
		utils.ErrorResponseWithData(
			c,
			http.StatusBadRequest,
			fmt.Sprintf("Page %d out of range. Only %d pages available", page, pagination.TotalPages),
			nil,
			"Page out of range",
		)
		return
	}

	var kebunList []models.Kebun
	if err := db.Limit(perPage).Offset(offset).Find(&kebunList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kebun", err.Error())
		return
	}

	if totalRows == 0 {
		utils.SuccessResponseWithMeta(c, http.StatusOK, "Tidak ada data kebun ditemukan", []models.Kebun{}, pagination)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Data kebun berhasil diambil", kebunList, pagination)
}

// GetKebunByID godoc
// @Summary     Ambil detail kebun
// @Description Mengambil detail kebun berdasarkan ID
// @Tags        Kebun
// @Param       id   path   int   true  "ID Kebun"
// @Success     200  {object} utils.Response
// @Failure     404  {object} utils.Response
// @Router      /kebun/{id} [get]
// GET /kebun/:id
func GetKebunByID(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var kebun models.Kebun
	if err := db.First(&kebun, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Kebun tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data kebun", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail kebun berhasil diambil", kebun)
}

// POST /kebun
// CreateKebun godoc
// @Summary     Membuat kebun baru
// @Description Membuat data kebun baru (response mengikuti utils.Response)
// @Tags        Kebun
// @Accept      json
// @Produce     json
// @Param       request  body   object{nama_kebun=string,mdpl=string}  true  "Input kebun"
// @Security 	Bearer
// @Success     201  {object} utils.Response
// @Failure     400  {object} utils.Response
// @Router      /kebun [post]
func CreateKebun(c *gin.Context) {
	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var input struct {
		NamaKebun string `json:"nama_kebun" binding:"required"`
		MDPL      string `json:"mdpl" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	nama := strings.TrimSpace(input.NamaKebun)
	if nama == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "nama_kebun tidak boleh kosong", nil)
		return
	}

	kebun := models.Kebun{
		NamaKebun: nama,
		MDPL:      input.MDPL,
	}

	if err := db.Create(&kebun).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat kebun", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Kebun berhasil dibuat", kebun)
}

// PUT /kebun/:id
// UpdateKebun godoc
// @Summary     Update kebun
// @Description Mengupdate data kebun berdasarkan ID (partial update)
// @Tags        Kebun
// @Accept      json
// @Produce     json
// @Param       id       path   int  true  "ID Kebun"
// @Param request body controllers.UpdateKebunRequest true "Input update"
// @Security 	Bearer
// @Success     200  {object} utils.Response
// @Failure     404  {object} utils.Response
// @Router      /kebun/{id} [put]
func UpdateKebun(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var kebun models.Kebun
	if err := db.First(&kebun, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Kebun tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data kebun", err.Error())
		return
	}

	var input UpdateKebunRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}

	if input.NamaKebun != nil {
		nama := strings.TrimSpace(*input.NamaKebun)
		if nama == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "nama_kebun tidak boleh kosong", nil)
			return
		}
		kebun.NamaKebun = nama
	}

	if input.MDPL != nil {
		kebun.MDPL = *input.MDPL
	}

	if err := db.Save(&kebun).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update kebun", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Kebun berhasil diperbarui", kebun)
}

// DELETE /kebun/:id
// DeleteKebun godoc
// @Summary     Hapus kebun
// @Description Menghapus data kebun berdasarkan ID
// @Tags        Kebun
// @Param       id   path   int  true  "ID Kebun"
// @Security 	Bearer
// @Success     200  {object} utils.Response
// @Failure     404  {object} utils.Response
// @Router      /kebun/{id} [delete]
func DeleteKebun(c *gin.Context) {
	id := c.Param("id")

	db, err := config.DbConnect()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
		return
	}

	var kebun models.Kebun
	if err := db.First(&kebun, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Kebun tidak ditemukan", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data kebun", err.Error())
		return
	}

	if err := db.Delete(&kebun).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal hapus kebun", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Kebun berhasil dihapus", utils.EmptyObj{})
}
