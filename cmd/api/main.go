package main

import (
	"context"
	"log"
	"os"
	"time"

	"hotel-app/config"
	"hotel-app/internal/auth"
	"hotel-app/internal/expense"
	"hotel-app/internal/notification"
	"hotel-app/internal/reservation"
	"hotel-app/internal/room"
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
		if err := config.AutoMigrate(
			&auth.User{},
			&room.Room{},
			&reservation.Reservation{},
			&expense.Expense{},
		); err != nil {
			log.Printf("Peringatan Migrasi: %v\n", err)
		} else {
			log.Println("Migrasi tabel users, rooms, reservations, dan expenses berhasil.")
		}
	}

	// Setup seluruh rute aplikasi
	r := routes.SetupRouter()

	// Jalankan background worker reminder checkout 15 menit (cek berkala setiap 1 menit)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifyService := notification.NewService()
	reservationRepo := reservation.NewRepository(config.DB)
	reservationService := reservation.NewService(reservationRepo, notifyService)
	reservation.StartReminderWorker(ctx, reservationService, 1*time.Minute)

	// Membaca port dari environment variable (default 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("Aplikasi Akuntansi Hotel berjalan di http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
