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

// RegisterRequest struct for registration request body
type RegisterRequest struct {
    FullName string `json:"full_name" example:"John Doe"`
    Email    string `json:"email" example:"john@example.com"`
    Phone    string `json:"phone" example:"08123456789"`
    Password string `json:"password" example:"secret123"`
}

// LoginRequest struct for login request body
type LoginRequest struct {
    Email    string `json:"email" example:"john@example.com"`
    Password string `json:"password" example:"secret123"`
}

// ManualRegisterPetani godoc
// @Summary Register Petani
// @Description Register petani menggunakan credential lokal
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Register Petani"
// @Success 201 {object} map[string]interface{} "Registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 409 {object} map[string]interface{} "Email or phone already exists"
// @Router /register/petani [post]
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

// ManualRegisterPembeli godoc
// @Summary Register Pembeli
// @Description Register pembeli menggunakan credential lokal
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Register Pembeli"
// @Success 201 {object} map[string]interface{} "Registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 409 {object} map[string]interface{} "Email or phone already exists"
// @Router /register/pembeli [post]
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

// ManualLogin godoc
// @Summary Login user
// @Description Login user menggunakan email dan password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login success with JWT"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Wrong credentials"
// @Router /login [post]
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