package middleware

import (
	"net/http"
	"strings"

	"hotel-app/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware memvalidasi token JWT pada header Authorization (Bearer <token>).
// Jika valid, user_id dan user_email disimpan ke context request.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Akses ditolak: header Authorization tidak ditemukan",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Format header otentikasi harus: Bearer <token>",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token tidak valid atau telah kedaluwarsa",
			})
			c.Abort()
			return
		}

		// Set data pengguna ke context Gin untuk digunakan oleh controller/handler
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}
