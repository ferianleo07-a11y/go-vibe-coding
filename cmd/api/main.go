package main

import (
	"log"
	"os"

	"hotel-app/config"
	"hotel-app/models"
	"hotel-app/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Memuat variabel lingkungan dari file .env (jika ada)
	if err := godotenv.Load(); err != nil {
		log.Println("Info: File .env tidak ditemukan, menggunakan environment variable sistem.")
	}

	// Inisialisasi koneksi database PostgreSQL
	_, err := config.InitDB()
	if err != nil {
		log.Printf("Peringatan Database: %v\n", err)
	} else {
		// Jalankan AutoMigrate skema tabel
		if err := config.AutoMigrate(&models.User{}, &models.Room{}, &models.Reservation{}); err != nil {
			log.Printf("Peringatan Migrasi: %v\n", err)
		} else {
			log.Println("Migrasi tabel users, rooms, dan reservations berhasil.")
		}
	}

	// Setup seluruh rute aplikasi
	r := routes.SetupRouter()

	// Membaca port dari environment variable (default 8080)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Aplikasi Akuntansi Hotel berjalan di http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
