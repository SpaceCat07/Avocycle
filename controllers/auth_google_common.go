package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

// Dipanggil dari CallbackHandlerPembeli / CallbackHandlerPetani
// defaultRole hanya untuk info ke FE (preselect), role final ditentukan di completeGoogleWithRole.
func handleGoogleCallback(c *gin.Context, defaultRole string) {
	state := c.Query("state")
	code := c.Query("code")
	provider := "google"

	googleUser, _, err := config.Gocial.Handle(state, code)
	if err != nil {
		redirectFrontendWithError(c, "google_handle_error")
		return
	}

	db, err := config.DbConnect()
	if err != nil {
		redirectFrontendWithError(c, "db_connect_failed")
		return
	}

	// cek user lama (auth_provider = "Google")
	var user models.User
	db.Where("auth_provider = ? AND provider_id = ?", "Google", googleUser.ID).
		First(&user)

	if user.ID != 0 {
		// ========== USER LAMA → langsung kirim JWT ke FE ==========
		jwtToken, err := utils.GenerateJWT(&user)
		if err != nil {
			redirectFrontendWithError(c, "jwt_generate_failed")
			return
		}

		frontendCallback := os.Getenv("FRONTEND_GOOGLE_CALLBACK_URL")
		if frontendCallback == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":    "FRONTEND_GOOGLE_CALLBACK_URL not set",
				"jwtToken": jwtToken,
				"user":     user,
			})
			return
		}

		u, _ := url.Parse(frontendCallback)
		q := u.Query()
		q.Set("token", jwtToken)
		u.RawQuery = q.Encode()

		c.Redirect(http.StatusTemporaryRedirect, u.String())
		return
	}

	// ========== USER BARU → kirim tempToken ke FE untuk pilih role ==========
	tempToken, err := utils.GenerateTempGoogleToken(
		provider,
		googleUser.ID,
		googleUser.Email,
		googleUser.FullName,
	)
	if err != nil {
		redirectFrontendWithError(c, "temp_token_failed")
		return
	}

	chooseRoleURL := os.Getenv("FRONTEND_CHOOSE_ROLE_URL")
	if chooseRoleURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "FRONTEND_CHOOSE_ROLE_URL not set",
			"tempToken": tempToken,
		})
		return
	}

	u, _ := url.Parse(chooseRoleURL)
	q := u.Query()
	q.Set("tempToken", tempToken)
	q.Set("defaultRole", defaultRole) // optional untuk preselect di FE
	u.RawQuery = q.Encode()

	c.Redirect(http.StatusTemporaryRedirect, u.String())
}

// Body dari FE: { "tempToken": "..." }
type CompleteGoogleReq struct {
	TempToken string `json:"tempToken" binding:"required"`
}

// Helper umum, dipanggil oleh CompleteGooglePembeli / CompleteGooglePetani
func completeGoogleWithRole(c *gin.Context, role string) {
	var req CompleteGoogleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}

	claims, err := utils.ParseTempGoogleToken(req.TempToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_or_expired_temp_token"})
		return
	}

	db, err := config.DbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_connect_failed"})
		return
	}

	// handle kalau user sudah dibuat
	var user models.User
	db.Where("auth_provider = ? AND provider_id = ?", "Google", claims.ProviderID).
		First(&user)

	if user.ID == 0 {
		user = models.User{
			FullName:     claims.FullName,
			Email:        claims.Email,
			AuthProvider: "Google",
			ProviderID:   claims.ProviderID,
			Role:         role,
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create_user_failed"})
			return
		}
	}

	jwtToken, err := utils.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "jwt_generate_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   jwtToken,
		"user":    user,
	})
}

// helper error → redirect ke FE dengan ?error=...
func redirectFrontendWithError(c *gin.Context, errCode string) {
	frontendCallback := os.Getenv("FRONTEND_GOOGLE_CALLBACK_URL")
	if frontendCallback == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errCode})
		return
	}

	u, _ := url.Parse(frontendCallback)
	q := u.Query()
	q.Set("error", errCode)
	u.RawQuery = q.Encode()

	c.Redirect(http.StatusTemporaryRedirect, u.String())
}
