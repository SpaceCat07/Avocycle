package controllers

import (
	"Avocycle/models"
	"Avocycle/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Avocycle/config"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

// @Definitions
// --- Struktur Helper untuk Anotasi Swagger ---

// LogPenyakitTanamanCustom dibuat eksplisit agar Swag tidak mencari gorm.Model
type LogPenyakitTanamanCustom struct {
    ID                uint      `json:"ID" example:"1"`
    CreatedAt         time.Time `json:"CreatedAt" example:"2025-11-28T08:37:35.000000000+07:00"`
    UpdatedAt         time.Time `json:"UpdatedAt" example:"2025-11-28T08:37:35.000000000+07:00"`
    TanamanID         uint      `json:"tanaman_id" example:"1"`
    PenyakitID        uint      `json:"penyakit_id" example:"2"`
    Kondisi           string    `json:"kondisi" example:"Parah"`
    SaranPerawatan    string    `json:"saran_perawatan" example:"Tingkatkan sirkulasi udara."`
    Foto              string    `json:"foto" example:"https://cloudinary.com/url..."`
    FotoLogPenyakitID string    `json:"foto_log_penyakit_id" example:"public_id_abc"`
}

// ClassifyPenyakitData menggunakan struktur custom di atas.
type ClassifyPenyakitData struct {
    NamaPenyakit     string                   `json:"nama_penyakit" example:"Antraknosa"`
    Deskripsi        string                   `json:"deskripsi" example:"Penyakit jamur yang menyebabkan bercak hitam."`
    Kondisi          string                   `json:"kondisi" example:"Sedang"`
    SaranPerawatan   string                   `json:"saran_perawatan" example:"Aplikasikan fungisida berbahan dasar tembaga."`
    Log              LogPenyakitTanamanCustom `json:"log"` // <-- Ganti models.LogPenyakitTanaman
}

// SuccessResponseWrapper mereplikasi utils.Response untuk kasus sukses.
// Ini digunakan oleh anotasi @Success 200.
type SuccessResponseWrapper struct {
    Success bool                 `json:"success" example:"true"`
    Message string               `json:"message" example:"Klasifikasi berhasil"`
    Data    ClassifyPenyakitData `json:"data"`
    Meta    interface{}          `json:"meta,omitempty"`
}

// ErrorResponseWrapper mereplikasi utils.Response untuk kasus error.
// Ini digunakan oleh anotasi @Failure 400 dan 500.
type ErrorResponseWrapper struct {
    Success bool        `json:"success" example:"false"`
    Message string      `json:"message" example:"Proses klasifikasi atau upload gagal"`
    // UBAH DARI interface{} (any) KE string
    Error   string      `json:"error,omitempty" example:"Gagal parsing hasil Gemini: invalid character..."` 
    Data    interface{} `json:"data,omitempty"`
    Meta    interface{} `json:"meta,omitempty"` 
}

// @Summary Klasifikasi Penyakit Tanaman Alpukat
// @Description Menerima foto tanaman alpukat, mengklasifikasikan penyakit menggunakan Gemini AI, mengunggah foto, dan menyimpan log ke database.
// @Tags Petani & Admin (Deteksi)
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param id_tanaman path int true "ID Tanaman yang ingin diklasifikasi penyakitnya"
// @Param foto_tanaman formData file true "Foto daun atau bagian tanaman yang sakit (JPG, PNG, dll.)"
// @Success 200 {object} controllers.SuccessResponseWrapper "Klasifikasi berhasil" // <-- Menggunakan wrapper dari package controllers
// @Failure 400 {object} controllers.ErrorResponseWrapper "ID tanaman tidak valid atau Foto tanaman wajib diunggah" // <-- Menggunakan wrapper dari package controllers
// @Failure 500 {object} controllers.ErrorResponseWrapper "Gagal konek DB, Inisialisasi Gemini gagal, Proses klasifikasi atau upload gagal, atau Gagal menyimpan data" // <-- Menggunakan wrapper dari package controllers
// @Router /petamin/penyakit/{id_tanaman} [post]
func ClassifyPenyakit(c *gin.Context) {
	// ambil id tanaman dari parameter url
	idTanaman := c.Param("id_tanaman")

	// connect to database
	db, err := config.DbConnect()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal konek DB", err.Error())
        return
    }

	// parse string into uint
	tanamanId, err := strconv.Atoi(idTanaman)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tanaman tidak valid", err.Error())
        return
	}

	// load env file
	err = godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// load photo file
	file, err := c.FormFile("foto_tanaman")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Foto tanaman wajib diunggah", err.Error())
        return
	}

	// open the file
	src, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuka file", err.Error())
        return
	}

	// close if error
	defer src.Close()

	// read image bytes
	imageBytes, err := io.ReadAll(src)
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membaca file", err.Error())
        return
    }

	// Tentukan MIME type berdasarkan ekstensi file (abaikan header dari client)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".webp":
		contentType = "image/webp"
	case ".gif":
		contentType = "image/gif"
	default:
		// fallback aman ke JPEG
		contentType = "image/jpeg"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

    client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API")))
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Inisialisasi Gemini gagal", err.Error())
        return
    }
    defer client.Close()

	var(
		classifyResult struct {
            NamaPenyakit    string `json:"nama_penyakit"`
            Deskripsi       string `json:"deskripsi"`
            Kondisi         string `json:"kondisi"`
            SaranPerawatan  string `json:"saran_perawatan"`
        }
        uploadURL string
        uploadID  string
	)
	g, gctx := errgroup.WithContext(ctx)

	g.Go(
		func() error {
			model := client.GenerativeModel("gemini-2.5-flash")

			part := genai.Blob{
				MIMEType: contentType,
				Data: imageBytes,
			}

			resp, err := model.GenerateContent(gctx,
			part,
			genai.Text(`Analisis foto tanaman alpukat ini dan hasilkan JSON dengan format persis:
{
  "nama_penyakit": "<nama penyakit atau 'Tidak terdeteksi'>",
  "deskripsi": "<deskripsi penyakit>",
  "kondisi": "<Parah|Sedang|Ringan|Sembuh>",
  "saran_perawatan": "<saran singkat>"
}`),
		)
		if err != nil {
			return err
		}

		text := cleanupGeminiJSON(extractText(resp))

		// --- LOGGING TAMBAHAN UNTUK DEBUGGING ---
		fmt.Printf("DEBUG: Gemini Raw Text Result: \n%s\n", text)
		// --- END LOGGING ---

		// Cek apakah hasil yang dibersihkan terlihat seperti JSON
		if !strings.HasPrefix(text, "{") {
			// Jika tidak diawali kurung kurawal, ini adalah kegagalan format.
			// Kembalikan error dengan hasil teks mentah dari Gemini.
			return fmt.Errorf("Gemini tidak mengembalikan JSON yang valid. Output: %s", text)
		}

		if err := json.Unmarshal([]byte(text), &classifyResult); err != nil {
			return fmt.Errorf("Gagal parsing hasil Gemini: %w", err)
		}
		return nil
	})

	g.Go(
		func() error {
			url, publicID, err := utils.AsyncUploadOptionalImage(file, "logpenyakit")
			if err != nil {
				return err
			}
			uploadURL = url
			uploadID = publicID
			return nil
		})

	if err := g.Wait(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Proses klasifikasi atau upload gagal", err.Error())
		return
	}

	penyakit := models.PenyakitTanaman{
		NamaPenyakit: classifyResult.NamaPenyakit,
		Deskripsi: classifyResult.Deskripsi,
	}

	if err := db.FirstOrCreate(&penyakit, models.PenyakitTanaman{NamaPenyakit: classifyResult.NamaPenyakit}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan penyakit", err.Error())
        return
	}

	logPenyakit := models.LogPenyakitTanaman{
		TanamanID: uint(tanamanId),
		PenyakitID: penyakit.ID,
		Kondisi: classifyResult.Kondisi,
		SaranPerawatan: classifyResult.SaranPerawatan,
		Foto: uploadURL,
		FotoLogPenyakitID: uploadID,
	}
	if err := db.Create(&logPenyakit).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan log penyakit", err.Error())
        return
	}

	// ===============================================
	// === START: TAMBAHAN KODE UNTUK PRELOAD RELASI ===
	// ===============================================

	// Ambil kembali logPenyakit yang baru dibuat, sambil memuat relasi Tanaman dan Penyakit
	if err := db.
		Preload("Tanaman.Kebun").  // Preload Tanaman, dan di dalamnya Preload Kebun
		Preload("Penyakit").       // Preload Penyakit
		First(&logPenyakit, logPenyakit.ID).Error; err != nil {
			// Jika gagal preload, kita log errornya tapi mungkin tetap mengirim response tanpa data lengkap
			fmt.Println("Warning: Gagal preload relasi logPenyakit:", err)
			// Lanjutkan tanpa return error, agar response sukses tetap terkirim
	}

	utils.SuccessResponse(c, http.StatusOK, "Klasifikasi berhasil", gin.H{
		"nama_penyakit":    classifyResult.NamaPenyakit,
		"deskripsi":		classifyResult.Deskripsi,
        "kondisi":          classifyResult.Kondisi,
        "saran_perawatan":  classifyResult.SaranPerawatan,
        "log":              logPenyakit,	
	})
}

func cleanupGeminiJSON(raw string) string {
    raw = strings.TrimSpace(raw)
    raw = strings.Trim(raw, "`")
    raw = strings.TrimPrefix(raw, "json")
    return strings.TrimSpace(raw)
}

func extractText(resp *genai.GenerateContentResponse) string {
    if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
        return ""
    }
    var b strings.Builder
    for _, part := range resp.Candidates[0].Content.Parts {
        if txt, ok := part.(genai.Text); ok {
            b.WriteString(string(txt))
        }
    }
    return b.String()
}