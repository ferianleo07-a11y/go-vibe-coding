package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User merepresentasikan admin / pengguna pemilik akun akuntansi hotel.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HashPassword mengenkripsi kata sandi menggunakan algoritma bcrypt sebelum disimpan ke database.
func (u *User) HashPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

// CheckPassword membandingkan plain text kata sandi dengan hash bcrypt yang tersimpan.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
