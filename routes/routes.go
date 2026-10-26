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

		// Rute Reservasi & Check-in / Check-out (Terlindungi JWT)
		reservations := api.Group("/reservations", middlewares.AuthMiddleware())
		{
			reservations.GET("", controllers.GetReservations)
			reservations.GET("/:id", controllers.GetReservationByID)
			reservations.POST("", controllers.CreateReservation)
			reservations.PUT("/:id/checkout", controllers.CheckoutReservation)
			reservations.PUT("/:id/cancel", controllers.CancelReservation)
		}

		// Rute Biaya Operasional USALI (Terlindungi JWT)
		expenses := api.Group("/expenses", middlewares.AuthMiddleware())
		{
			expenses.GET("/categories", controllers.GetExpenseCategories)
			expenses.GET("", controllers.GetExpenses)
			expenses.GET("/:id", controllers.GetExpenseByID)
			expenses.POST("", controllers.CreateExpense)
			expenses.PUT("/:id", controllers.UpdateExpense)
			expenses.DELETE("/:id", controllers.DeleteExpense)
		}

		// Rute Dashboard KPI & Rekap Laporan Keuangan (Terlindungi JWT)
		reports := api.Group("/reports", middlewares.AuthMiddleware())
		{
			reports.GET("/dashboard", controllers.GetDashboardMetrics)
		}
	}

	return r
}
