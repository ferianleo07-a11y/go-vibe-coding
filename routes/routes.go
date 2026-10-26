package routes

import (
	"net/http"

	"hotel-app/config"
	"hotel-app/controllers"
	"hotel-app/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRouter mengonfigurasi seluruh endpoint REST API dan penyajian frontend statis.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 1. Health check endpoint
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

	// 2. Penyajian berkas statis frontend (HTML, CSS, JS)
	r.Static("/static", "./web")
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/login", "./web/login.html")
	r.StaticFile("/register", "./web/register.html")

	// 3. Grup endpoint API
	api := r.Group("/api")
	{
		// Autentikasi Publik
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)

			// Profil Pengguna (Terlindungi JWT)
			auth.GET("/me", middlewares.AuthMiddleware(), controllers.GetCurrentUser)
		}

		// Rute Kamar (Terlindungi JWT)
		rooms := api.Group("/rooms", middlewares.AuthMiddleware())
		{
			rooms.GET("", controllers.GetRooms)
			rooms.GET("/:id", controllers.GetRoomByID)
			rooms.POST("", controllers.CreateRoom)
			rooms.PUT("/:id", controllers.UpdateRoom)
			rooms.DELETE("/:id", controllers.DeleteRoom)
		}
	}

	return r
}
