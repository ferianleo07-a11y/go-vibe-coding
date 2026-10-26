package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes PostgreSQL connection using DATABASE_URL environment variable.
// It ensures credentials are never hardcoded and handles connection errors safely.
func InitDB() (*gorm.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("variabel lingkungan DATABASE_URL belum diatur. Silakan periksa file .env Anda")
	}

	// Open connection with GORM
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		// Log informative error without printing the full connection string/credentials
		log.Printf("Gagal terhubung ke database PostgreSQL. Pastikan database server aktif.")
		return nil, fmt.Errorf("koneksi database gagal: periksa status server PostgreSQL Anda")
	}

	DB = db
	log.Println("Berhasil terhubung ke database PostgreSQL.")
	return DB, nil
}

// AutoMigrate menjalankan migrasi skema tabel database untuk model-model yang diberikan.
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return errors.New("database belum terhubung, migrasi dilewati")
	}
	return DB.AutoMigrate(models...)
}
