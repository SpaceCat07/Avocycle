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