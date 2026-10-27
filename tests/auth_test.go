package tests

import (
	"testing"

	"hotel-app/internal/auth"
	"hotel-app/utils"
)

// TestPasswordHashing menguji keamanan hashing password menggunakan bcrypt
func TestPasswordHashing(t *testing.T) {
	user := auth.User{
		Name:  "Admin Hotel",
		Email: "admin@test.com",
	}

	password := "rahasia123"
	if err := user.HashPassword(password); err != nil {
		t.Fatalf("Gagal mengenkripsi kata sandi: %v", err)
	}

	// Pastikan password tidak tersimpan dalam bentuk plain text
	if user.Password == password {
		t.Fatalf("Kata sandi tidak boleh disimpan dalam format plain text")
	}

	// Uji verifikasi password yang benar
	if !user.CheckPassword(password) {
		t.Fatalf("Verifikasi password yang benar seharusnya berhasil")
	}

	// Uji verifikasi password yang salah
	if user.CheckPassword("passwordsalah") {
		t.Fatalf("Verifikasi password yang salah tidak boleh berhasil")
	}
}

// TestJWTTokenLifecycle menguji pembuatan dan validasi token JWT
func TestJWTTokenLifecycle(t *testing.T) {
	userID := uint(10)
	email := "finance@hotel.com"

	token, err := utils.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Gagal membuat token JWT: %v", err)
	}

	if token == "" {
		t.Fatalf("Token JWT yang dihasilkan tidak boleh kosong")
	}

	// Validasi token
	claims, err := utils.ValidateToken(token)
	if err != nil {
		t.Fatalf("Validasi token gagal: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Ekspektasi UserID %d, didapatkan %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Ekspektasi Email %s, didapatkan %s", email, claims.Email)
	}

	// Uji validasi token palsu
	_, err = utils.ValidateToken("invalid.token.payload")
	if err == nil {
		t.Fatalf("Token palsu seharusnya ditolak")
	}
}
