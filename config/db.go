package config

import (
	"Avocycle/models"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DbConnect() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

    // ini untuk migrate nanti
    db.AutoMigrate(
		&models.User{},
		&models.Kebun{},
		&models.ProsesProduksi{},
		&models.PerawatanPenyakit{},
		&models.PenyakitTanaman{},
		&models.Tanaman{},
		&models.LogPenyakitTanaman{},
		&models.Buah{},
		&models.LogProsesProduksi{},
		&models.Booking{},
		&models.PersonalAccessTokens{},
	)

	return db, nil
}