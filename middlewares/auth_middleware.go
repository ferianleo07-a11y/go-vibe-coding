package middlewares

import (
	"hotel-app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware memvalidasi token JWT pada header Authorization.
func AuthMiddleware() gin.HandlerFunc {
	return middleware.AuthMiddleware()
}
