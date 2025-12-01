// controllers/auth_google_complete.go
package controllers

import "github.com/gin-gonic/gin"

// POST /api/v1/auth/google/complete/pembeli
func CompleteGooglePembeli(c *gin.Context) {
    completeGoogleWithRole(c, "Pembeli")
}

// POST /api/v1/auth/google/complete/petani
func CompleteGooglePetani(c *gin.Context) {
    completeGoogleWithRole(c, "Petani")
}
