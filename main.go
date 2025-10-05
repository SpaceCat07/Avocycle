package main

import (
	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/routes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// === 1️⃣ Connect ke PostgreSQL ===
	db, err := config.DbConnect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get underlying DB instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("✅ Connected to PostgreSQL database (LARAGON)")
	fmt.Println("========================================")

	// === 2️⃣ Auto migrate semua model ===
	if err := db.AutoMigrate(
		&models.User{},
		&models.Tanaman{},
		&models.ProsesProduksi{},
		&models.PersonalAccessTokens{},
		&models.PerawatanPenyakit{},
		&models.PenyakitTanaman{},
		&models.LogProsesProduksi{},
		&models.LogPenyakitTanaman{},
		&models.Kebun{},
		&models.Buah{},
		&models.Booking{},
	); err != nil {
		log.Fatalf("❌ Failed to auto migrate models: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("✅ Database migration completed successfully")
	fmt.Println("========================================")

	// === 3️⃣ Inisialisasi routes ===
	router := routes.InitRoutes()

	// === 4️⃣ Jalankan server (dengan graceful shutdown) ===
	srv := &http.Server{
		Addr:    ":2005",
		Handler: router,
	}

	// Jalankan server di goroutine agar bisa ditutup dengan graceful shutdown
	go func() {
		fmt.Println("🚀 Server is running on http://localhost:2005")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// === 5️⃣ Graceful shutdown (Ctrl+C atau kill signal) ===
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // tunggu sinyal berhenti

	fmt.Println("\n🛑 Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("⚠️  Failed to close DB connection: %v", err)
	} else {
		fmt.Println("✅ Database connection closed.")
	}

	fmt.Println("👋 Server stopped cleanly.")
}
