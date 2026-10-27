package routes

import (
	"net/http"

	"hotel-app/config"
	"hotel-app/internal/auth"
	"hotel-app/internal/expense"
	"hotel-app/internal/middleware"
	"hotel-app/internal/notification"
	"hotel-app/internal/report"
	"hotel-app/internal/reservation"
	"hotel-app/internal/room"

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

	// Dependency Injection untuk masing-masing domain/fitur
	dbReadyFunc := func() bool { return config.DB != nil }

	// Notification Service (WhatsApp & Email Reminder)
	notifyService := notification.NewService()

	// Auth Domain
	authRepo := auth.NewRepository(config.DB)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService, dbReadyFunc)

	// Room Domain
	roomRepo := room.NewRepository(config.DB)
	roomService := room.NewService(roomRepo)
	roomHandler := room.NewHandler(roomService, dbReadyFunc)

	// Reservation Domain
	reservationRepo := reservation.NewRepository(config.DB)
	reservationService := reservation.NewService(reservationRepo, notifyService)
	reservationHandler := reservation.NewHandler(reservationService, dbReadyFunc)

	// Expense Domain
	expenseRepo := expense.NewRepository(config.DB)
	expenseService := expense.NewService(expenseRepo)
	expenseHandler := expense.NewHandler(expenseService, dbReadyFunc)

	// Report Domain
	reportRepo := report.NewRepository(config.DB)
	reportService := report.NewService(reportRepo)
	reportHandler := report.NewHandler(reportService, dbReadyFunc)

	// 3. Grup endpoint API
	api := r.Group("/api")
	{
		// Autentikasi Publik
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)

			// Profil Pengguna (Terlindungi JWT)
			authGroup.GET("/me", middleware.AuthMiddleware(), authHandler.GetCurrentUser)
		}

		// Rute Kamar (Terlindungi JWT)
		roomsGroup := api.Group("/rooms", middleware.AuthMiddleware())
		{
			roomsGroup.GET("", roomHandler.GetRooms)
			roomsGroup.GET("/:id", roomHandler.GetRoomByID)
			roomsGroup.POST("", roomHandler.CreateRoom)
			roomsGroup.PUT("/:id", roomHandler.UpdateRoom)
			roomsGroup.DELETE("/:id", roomHandler.DeleteRoom)
		}

		// Rute Reservasi & Check-in / Check-out (Terlindungi JWT)
		reservationsGroup := api.Group("/reservations", middleware.AuthMiddleware())
		{
			reservationsGroup.GET("", reservationHandler.GetReservations)
			reservationsGroup.GET("/payment-methods", reservationHandler.GetPaymentMethods)
			reservationsGroup.GET("/:id", reservationHandler.GetReservationByID)
			reservationsGroup.POST("", reservationHandler.CreateReservation)
			reservationsGroup.PUT("/:id/checkout", reservationHandler.CheckoutReservation)
			reservationsGroup.PUT("/:id/cancel", reservationHandler.CancelReservation)
			reservationsGroup.POST("/:id/remind", reservationHandler.SendReminder)
		}

		// Rute Biaya Operasional USALI (Terlindungi JWT)
		expensesGroup := api.Group("/expenses", middleware.AuthMiddleware())
		{
			expensesGroup.GET("/categories", expenseHandler.GetExpenseCategories)
			expensesGroup.GET("", expenseHandler.GetExpenses)
			expensesGroup.GET("/:id", expenseHandler.GetExpenseByID)
			expensesGroup.POST("", expenseHandler.CreateExpense)
			expensesGroup.PUT("/:id", expenseHandler.UpdateExpense)
			expensesGroup.DELETE("/:id", expenseHandler.DeleteExpense)
		}

		// Rute Dashboard KPI & Rekap Laporan Keuangan (Terlindungi JWT)
		reportsGroup := api.Group("/reports", middleware.AuthMiddleware())
		{
			reportsGroup.GET("/dashboard", reportHandler.GetDashboardMetrics)
		}
	}

	return r
}
