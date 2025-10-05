package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ===================================================
// 🔐 Utility Function — Register User per Role
// ===================================================
func registerUser(c *gin.Context, role string) {
	var req struct {
		FullName        string `json:"name" binding:"required"`
		Email           string `json:"email" binding:"required"`
		Phone           string `json:"phone" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid or missing fields in JSON body",
			"details": err.Error(),
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid email format"})
		return
	}

	phoneRegex := regexp.MustCompile(`^[0-9]{10,15}$`)
	if !phoneRegex.MatchString(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Phone number must contain only digits and be between 10–15 digits"})
		return
	}

	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Password must be at least 8 characters long"})
		return
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range req.Password {
		switch {
		case 'a' <= ch && ch <= 'z':
			hasLower = true
		case 'A' <= ch && ch <= 'Z':
			hasUpper = true
		case '0' <= ch && ch <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Password must contain uppercase, lowercase, and a number"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Password and confirm password do not match"})
		return
	}

	db, err := config.DbConnect()
	if err != nil {
		log.Printf("[ERROR] Database connection failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to database"})
		return
	}

	var existingUser models.User
	err = db.Where("email = ?", email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "Email already registered"})
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[ERROR] Database query error during registration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database query error"})
		return
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to hash password"})
		return
	}

	user := models.User{
		FullName:     req.FullName,
		Email:        email,
		Phone:        req.Phone,
		PasswordHash: string(hashedPass),
		Role:         role,
		AuthProvider: "Local",
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("[ERROR] Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create user"})
		return
	}

	// ✅ Tambahkan pembuatan token di sini
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("[ERROR] Missing JWT_SECRET in .env")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "JWT secret not configured"})
		return
	}

	duration := 24 // default 24 jam
	if durStr := os.Getenv("JWT_EXPIRATION_HOURS"); durStr != "" {
		if d, err := strconv.Atoi(durStr); err == nil {
			duration = d
		}
	}

	claims := jwt.MapClaims{
		"sub":     user.Email,
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(duration) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     "avocycle-api",
		"aud":     "avocycle-frontend",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("[ERROR] Failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to generate token"})
		return
	}

	// ✅ Log success + kirim token langsung
	log.Printf("[INFO] User registered and logged in: %s (%s)", user.Email, user.Role)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"action":  "Register and auto-login for " + role,
		"data": gin.H{
			"id":        user.ID,
			"name":      user.FullName,
			"email":     user.Email,
			"phone":     user.Phone,
			"role":      user.Role,
			"token":     tokenString,
			"createdAt": user.CreatedAt,
		},
	})
}

// ===================================================
// 🔐 Endpoint Register Per Role
// ===================================================
func RegisterPetani(c *gin.Context)  { registerUser(c, "Petani") }
func RegisterPembeli(c *gin.Context) { registerUser(c, "Pembeli") }
func RegisterAdmin(c *gin.Context)   { registerUser(c, "Admin") }

// ===================================================
// 🔑 LOGIN HANDLER
// ===================================================
func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// 1️⃣ Validasi body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body", "details": err.Error()})
		return
	}

	// 2️⃣ Koneksi DB
	db, err := config.DbConnect()
	if err != nil {
		log.Printf("[ERROR] Database connection failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to database"})
		return
	}

	// 3️⃣ Cari user
	var user models.User
	if err := db.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
			return
		}
		log.Printf("[ERROR] DB query error on login: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database query error"})
		return
	}

	// 4️⃣ Rate limiting sederhana
	if user.FailedLoginAttempts >= 5 && time.Since(user.LastLoginAttempt) < 15*time.Minute {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "Too many failed login attempts. Try again later.",
		})
		return
	}

	// 5️⃣ Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		user.FailedLoginAttempts++
		user.LastLoginAttempt = time.Now()
		db.Save(&user)

		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid credentials"})
		return
	}

	// 6️⃣ Reset percobaan login gagal
	user.FailedLoginAttempts = 0
	db.Save(&user)

	// 7️⃣ Ambil secret dari ENV
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("[ERROR] Missing JWT_SECRET in .env")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "JWT secret not configured"})
		return
	}

	// 8️⃣ Ambil durasi token dari ENV
	duration := 24 // default 24 jam
	if durStr := os.Getenv("JWT_EXPIRATION_HOURS"); durStr != "" {
		if d, err := strconv.Atoi(durStr); err == nil {
			duration = d
		}
	}

	// 9️⃣ Buat JWT
	claims := jwt.MapClaims{
		"sub":     user.Email,
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(duration) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     "avocycle-api",
		"aud":     "avocycle-frontend",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("[ERROR] Failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to generate token"})
		return
	}

	// 🔟 Log aktivitas login
	log.Printf("[INFO] User login success: %s (%s)", user.Email, user.Role)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": gin.H{
			"id":    user.ID,
			"name":  user.FullName,
			"email": user.Email,
			"role":  user.Role,
			"token": tokenString,
		},
	})
}
