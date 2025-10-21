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

func ManualRegisterPembeli(c *gin.Context) {
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
    pembeli := models.User{
        FullName:     strings.TrimSpace(requestBody.FullName),
        Email:        strings.ToLower(strings.TrimSpace(requestBody.Email)),
        Phone:        strings.TrimSpace(requestBody.Phone),
        PasswordHash: string(hashedpass),
        Role:         "Pembeli",
        AuthProvider: "Local",
    }

	// Create user in database
    if err := db.Create(&pembeli).Error; err != nil {
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
        "message": "Pembeli registered successfully",
        "data": gin.H{
            "id":            pembeli.ID,
            "full_name":     pembeli.FullName,
            "email":         pembeli.Email,
            "phone":         pembeli.Phone,
            "role":          pembeli.Role,
            "auth_provider": pembeli.AuthProvider,
            "created_at":    pembeli.CreatedAt,
        },
    })
}

func ManualLogin(c *gin.Context) {
	var requestBody struct {
		Email string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
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

	// Get database connection
    db, err := config.DbConnect()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
        return
    }

	var user models.User

	// Find user by email
    if err := db.Where("email = ?", requestBody.Email).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Invalid email or password",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "success": false,
                "error":   "Database error",
            })
        }
        return
    }

	// check user registered via local not google
	if user.AuthProvider != "Local" {
		c.JSON(http.StatusNotAcceptable, gin.H{"error" : "Failed to load"})
		return
	}

	// compare password
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(requestBody.Password)) ; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error" : "Invalid Username Or Password"})
		return 
	}

	// generate jwt token
	token, err := utils.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "Failed to Generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Login successful",
        "data": gin.H{
            "id":            user.ID,
            "full_name":     user.FullName,
            "email":         user.Email,
            "phone":         user.Phone,
            "role":          user.Role,
            "auth_provider": user.AuthProvider,
        },
        "token": token,
    })
}