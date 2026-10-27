package report

import (
	"math"
)

// Service mendefinisikan antarmuka kalkulasi logika bisnis KPI hotel dan laporan keuangan.
type Service interface {
	GetDashboardMetrics(userID uint) (*DashboardMetricsResponse, error)
}

type service struct {
	repo Repository
}

// NewService membuat instance baru dari Report Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetDashboardMetrics(userID uint) (*DashboardMetricsResponse, error) {
	counts, err := s.repo.GetRoomCounts(userID)
	if err != nil {
		return nil, err
	}

	revenueStats, err := s.repo.GetRevenueStats(userID)
	if err != nil {
		return nil, err
	}

	totalExpenses, err := s.repo.GetTotalExpenses(userID)
	if err != nil {
		return nil, err
	}

	breakdown, err := s.repo.GetExpensesByCategory(userID)
	if err != nil {
		return nil, err
	}

	// Kalkulasi Metrik Indikator Utama Perhotelan (KPI Hotel)
	var occupancyRate float64 = 0
	if counts.Total > 0 {
		// Occupancy Rate (%) = (Kamar Terisi / Total Kamar) * 100
		occupancyRate = (float64(counts.Occupied) / float64(counts.Total)) * 100
		occupancyRate = math.Round(occupancyRate*100) / 100
	}

	var adr float64 = 0
	if revenueStats.SoldReservationsCount > 0 {
		// ADR = Total Pendapatan Kamar / Jumlah Transaksi Kamar Terjual
		adr = revenueStats.TotalRevenue / float64(revenueStats.SoldReservationsCount)
		adr = math.Round(adr)
	}

	var revpar float64 = 0
	if counts.Total > 0 {
		// RevPAR = Total Pendapatan Kamar / Total Kamar Tersedia
		revpar = revenueStats.TotalRevenue / float64(counts.Total)
		revpar = math.Round(revpar)
	}

	// Laba / Rugi Bersih (Net Income) = Total Pendapatan - Total Pengeluaran
	netIncome := revenueStats.TotalRevenue - totalExpenses

	response := &DashboardMetricsResponse{
		TotalRooms:         counts.Total,
		AvailableRooms:     counts.Available,
		OccupiedRooms:      counts.Occupied,
		CleaningRooms:      counts.Cleaning,
		MaintenanceRooms:   counts.Maintenance,
		OccupancyRate:      occupancyRate,
		RoomRevenue:        revenueStats.TotalRevenue,
		ADR:                adr,
		RevPAR:             revpar,
		TotalExpenses:      totalExpenses,
		NetIncome:          netIncome,
		ExpensesByCategory: breakdown,
	}

	return response, nil
}
