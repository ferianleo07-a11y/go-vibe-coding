package main

import (
	"log"
	"net/http"
	"os"

	"hotel-app/config"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Memuat variabel lingkungan dari file .env (jika ada)
	if err := godotenv.Load(); err != nil {
		log.Println("Info: File .env tidak ditemukan, menggunakan environment variable sistem.")
	}

	// Inisialisasi koneksi database PostgreSQL
	db, err := config.InitDB()
	if err != nil {
		log.Printf("Peringatan Database: %v\n", err)
	} else {
		_ = db
	}

	// Inisialisasi router Gin
	r := gin.Default()

	// Endpoint dasar health check untuk pengujian Tahap 1
	r.GET("/api/health", func(c *gin.Context) {
		dbStatus := "terhubung"
		if config.DB == nil {
			dbStatus = "tidak terhubung"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"app":      "Aplikasi Akuntansi & Keuangan Bisnis Hotel",
			"database": dbStatus,
		})
	})

	// Membaca port dari environment variable (default 8080)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
