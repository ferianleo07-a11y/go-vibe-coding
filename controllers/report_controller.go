package controllers

import (
	"math"
	"net/http"

	"hotel-app/config"
	"hotel-app/models"

	"github.com/gin-gonic/gin"
)

type CategoryExpenseBreakdown struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

type DashboardMetricsResponse struct {
	TotalRooms        int64                      `json:"total_rooms"`
	AvailableRooms    int64                      `json:"available_rooms"`
	OccupiedRooms     int64                      `json:"occupied_rooms"`
	CleaningRooms     int64                      `json:"cleaning_rooms"`
	MaintenanceRooms  int64                      `json:"maintenance_rooms"`
	OccupancyRate     float64                    `json:"occupancy_rate"` // %
	RoomRevenue       float64                    `json:"room_revenue"`
	ADR               float64                    `json:"adr"`    // Average Daily Rate
	RevPAR            float64                    `json:"revpar"` // Revenue per Available Room
	TotalExpenses     float64                    `json:"total_expenses"`
	NetIncome         float64                    `json:"net_income"` // Laba / Rugi Bersih
	ExpensesByCategory []CategoryExpenseBreakdown `json:"expenses_by_category"`
}

// GetDashboardMetrics menghitung metrik KPI hotel (Occupancy, ADR, RevPAR) dan rekapitulasi Laba/Rugi standar USALI.
func GetDashboardMetrics(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	// 1. Hitung jumlah kamar dan status ketersediaan
	var totalRooms int64
	var availableRooms int64
	var occupiedRooms int64
	var cleaningRooms int64
	var maintenanceRooms int64

	config.DB.Model(&models.Room{}).Where("user_id = ?", userID).Count(&totalRooms)
	config.DB.Model(&models.Room{}).Where("user_id = ? AND status = ?", userID, models.RoomStatusAvailable).Count(&availableRooms)
	config.DB.Model(&models.Room{}).Where("user_id = ? AND status = ?", userID, models.RoomStatusOccupied).Count(&occupiedRooms)
	config.DB.Model(&models.Room{}).Where("user_id = ? AND status = ?", userID, models.RoomStatusCleaning).Count(&cleaningRooms)
	config.DB.Model(&models.Room{}).Where("user_id = ? AND status = ?", userID, models.RoomStatusMaintenance).Count(&maintenanceRooms)

	// 2. Hitung total pendapatan kamar dari transaksi non-batal
	var totalRoomRevenue float64
	var soldReservationsCount int64

	config.DB.Model(&models.Reservation{}).
		Where("user_id = ? AND status != ?", userID, models.ReservationStatusCancelled).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRoomRevenue)

	config.DB.Model(&models.Reservation{}).
		Where("user_id = ? AND status != ?", userID, models.ReservationStatusCancelled).
		Count(&soldReservationsCount)

	// 3. Hitung total pengeluaran operasional USALI
	var totalExpenses float64
	config.DB.Model(&models.Expense{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpenses)

	// 4. Breakdown pengeluaran per kategori USALI
	var breakdown []CategoryExpenseBreakdown
	config.DB.Model(&models.Expense{}).
		Select("category, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ?", userID).
		Group("category").
		Order("total desc").
		Scan(&breakdown)

	// 5. Kalkulasi Metrik Indikator Utama Perhotelan (KPI Hotel)
	var occupancyRate float64 = 0
	if totalRooms > 0 {
		// Occupancy Rate (%) = (Kamar Terisi / Total Kamar) * 100
		occupancyRate = (float64(occupiedRooms) / float64(totalRooms)) * 100
		occupancyRate = math.Round(occupancyRate*100) / 100
	}

	var adr float64 = 0
	if soldReservationsCount > 0 {
		// ADR = Total Pendapatan Kamar / Jumlah Transaksi Kamar Terjual
		adr = totalRoomRevenue / float64(soldReservationsCount)
		adr = math.Round(adr)
	}

	var revpar float64 = 0
	if totalRooms > 0 {
		// RevPAR = Total Pendapatan Kamar / Total Kamar Tersedia
		revpar = totalRoomRevenue / float64(totalRooms)
		revpar = math.Round(revpar)
	}

	// Laba / Rugi Bersih (Net Income) = Total Pendapatan - Total Pengeluaran
	netIncome := totalRoomRevenue - totalExpenses

	response := DashboardMetricsResponse{
		TotalRooms:         totalRooms,
		AvailableRooms:     availableRooms,
		OccupiedRooms:      occupiedRooms,
		CleaningRooms:      cleaningRooms,
		MaintenanceRooms:   maintenanceRooms,
		OccupancyRate:      occupancyRate,
		RoomRevenue:        totalRoomRevenue,
		ADR:                adr,
		RevPAR:             revpar,
		TotalExpenses:      totalExpenses,
		NetIncome:          netIncome,
		ExpensesByCategory: breakdown,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}
