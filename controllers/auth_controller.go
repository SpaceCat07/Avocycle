package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"net/http"
	// "strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	// "gorm.io/gorm"
)

func ManualRegisterPetani(c *gin.Context) {
	var requestBody struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
        Phone string `json:"phone"`
        Password string `json:"password"`
	}

	// Get database connection
    db, err := config.DbConnect()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
        return
    }

	if err := c.ShouldBindJSON(&requestBody) ; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Invalid JSON format"})
		return
	}

	hashedpass, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	petani := models.User{
		FullName: string(requestBody.FullName),
		Email: string(requestBody.Email),
		Phone: string(requestBody.Phone),
		PasswordHash: string(hashedpass),
		Role: string("Petani"),
		AuthProvider: string("Manual"),
	}

	if err := db.Create(&petani).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "Failed to create User"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success" : true,
		"action" : "Register manual for petani",
		"data" : petani,
	})
}