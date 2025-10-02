package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
	"net/http"
	"strings"


	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ManualRegisterPetani(c *gin.Context) {
	var requestBody struct {
        FullName string `json:"full_name" binding:"required"`
        Email    string `json:"email" binding:"required,email"`
        Phone    string `json:"phone" binding:"required"`
        Password string `json:"password" binding:"required,min=6"`
    }

	// Get database connection
    db, err := config.DbConnect()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
        return
    }

	// Bind and validate JSON
    if err := c.ShouldBindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "Invalid input data",
            "details": err.Error(),
        })
        return
    }

	// Additional custom validations
    if err := utils.ValidateUserInput(requestBody.FullName, requestBody.Email, requestBody.Phone, requestBody.Password); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   err.Error(),
        })
        return
    }

	// Check if email already exists
    var existingUser models.User
    if err := db.Where("email = ?", strings.ToLower(requestBody.Email)).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{
            "success": false,
            "error":   "Email already registered",
        })
        return
    } else if err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Failed to check existing user",
        })
        return
    }

    // Check if phone already exists
    if err := db.Where("phone = ?", requestBody.Phone).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{
            "success": false,
            "error":   "Phone number already registered",
        })
        return
    } else if err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Failed to check existing phone",
        })
        return
    }

	hashedpass, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user object
    petani := models.User{
        FullName:     strings.TrimSpace(requestBody.FullName),
        Email:        strings.ToLower(strings.TrimSpace(requestBody.Email)),
        Phone:        strings.TrimSpace(requestBody.Phone),
        PasswordHash: string(hashedpass),
        Role:         "Petani",
        AuthProvider: "Local",
    }

	// Create user in database
    if err := db.Create(&petani).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Failed to create user",
            "details": err.Error(),
        })
        return
    }

	// Return success response (exclude password hash)
    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "message": "Petani registered successfully",
        "data": gin.H{
            "id":            petani.ID,
            "full_name":     petani.FullName,
            "email":         petani.Email,
            "phone":         petani.Phone,
            "role":          petani.Role,
            "auth_provider": petani.AuthProvider,
            "created_at":    petani.CreatedAt,
        },
    })
}