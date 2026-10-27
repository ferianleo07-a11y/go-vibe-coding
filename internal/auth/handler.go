package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler menangani request HTTP untuk domain otentikasi.
type Handler struct {
	service Service
	dbReady func() bool
}

// NewHandler membuat instance baru dari Auth Handler.
func NewHandler(service Service, dbReady func() bool) *Handler {
	return &Handler{
		service: service,
		dbReady: dbReady,
	}
}

// Register menangani pembuatan akun admin/keuangan baru.
func (h *Handler) Register(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan database saat ini tidak tersedia"})
		return
	}

	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: pastikan nama, email valid, dan password minimal 6 karakter"})
		return
	}

	userResp, err := h.service.Register(input)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pendaftaran berhasil, silakan masuk",
		"user": gin.H{
			"id":         userResp.ID,
			"name":       userResp.Name,
			"email":      userResp.Email,
			"created_at": userResp.CreatedAt,
		},
	})
}

// Login memvalidasi kredensial pengguna dan mengembalikan JWT token.
func (h *Handler) Login(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan database saat ini tidak tersedia"})
		return
	}

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan password wajib diisi"})
		return
	}

	token, userResp, err := h.service.Login(input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   token,
		"user": gin.H{
			"id":    userResp.ID,
			"name":  userResp.Name,
			"email": userResp.Email,
		},
	})
}

// GetCurrentUser mengembalikan profil pengguna yang sedang login berdasarkan token JWT.
func (h *Handler) GetCurrentUser(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan database saat ini tidak tersedia"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak ditemukan"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	userResp, err := h.service.GetCurrentUser(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         userResp.ID,
			"name":       userResp.Name,
			"email":      userResp.Email,
			"created_at": userResp.CreatedAt,
		},
	})
}
