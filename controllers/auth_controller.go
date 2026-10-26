package controllers

import (
	"net/http"

	"hotel-app/config"
	"hotel-app/models"
	"hotel-app/utils"

	"github.com/gin-gonic/gin"
)

// RegisterInput mendefinisikan validasi payload pendaftaran akun baru.
type RegisterInput struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginInput mendefinisikan validasi payload masuk akun.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register menangani pembuatan akun admin/keuangan baru.
func Register(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan database saat ini tidak tersedia"})
		return
	}

	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: pastikan nama, email valid, dan password minimal 6 karakter"})
		return
	}

	// Cek apakah email sudah terdaftar
	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email sudah digunakan oleh akun lain"})
		return
	}

	user := models.User{
		Name:  input.Name,
		Email: input.Email,
	}

	if err := user.HashPassword(input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses keamanan kata sandi"})
		return
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan akun ke database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pendaftaran berhasil, silakan masuk",
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})
}

// Login memvalidasi kredensial pengguna dan mengembalikan JWT token.
func Login(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan database saat ini tidak tersedia"})
		return
	}

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan password wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password tidak sesuai"})
		return
	}

	if !user.CheckPassword(input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password tidak sesuai"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menerbitkan token otentikasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// GetCurrentUser mengembalikan profil pengguna yang sedang login berdasarkan token JWT.
func GetCurrentUser(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan database saat ini tidak tersedia"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak ditemukan"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengguna tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})
}
