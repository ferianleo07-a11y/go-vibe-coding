package report

import (
	"hotel-app/config"

	"gorm.io/gorm"
)

type RoomCounts struct {
	Total       int64
	Available   int64
	Occupied    int64
	Cleaning    int64
	Maintenance int64
}

type RevenueStats struct {
	TotalRevenue          float64
	SoldReservationsCount int64
}

// Repository mendefinisikan antarmuka akses data statistik dan metrik keuangan.
type Repository interface {
	GetRoomCounts(userID uint) (RoomCounts, error)
	GetRevenueStats(userID uint) (RevenueStats, error)
	GetTotalExpenses(userID uint) (float64, error)
	GetExpensesByCategory(userID uint) ([]CategoryExpenseBreakdown, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository membuat instance baru dari Report Repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return config.DB
}

func (r *repository) GetRoomCounts(userID uint) (RoomCounts, error) {
	var counts RoomCounts

	if err := r.getDB().Table("rooms").Where("user_id = ?", userID).Count(&counts.Total).Error; err != nil {
		return counts, err
	}
	r.getDB().Table("rooms").Where("user_id = ? AND status = ?", userID, "Tersedia").Count(&counts.Available)
	r.getDB().Table("rooms").Where("user_id = ? AND status = ?", userID, "Terisi").Count(&counts.Occupied)
	r.getDB().Table("rooms").Where("user_id = ? AND status = ?", userID, "Sedang Dibersihkan").Count(&counts.Cleaning)
	r.getDB().Table("rooms").Where("user_id = ? AND status = ?", userID, "Dalam Perbaikan").Count(&counts.Maintenance)

	return counts, nil
}

func (r *repository) GetRevenueStats(userID uint) (RevenueStats, error) {
	var stats RevenueStats

	err := r.getDB().Table("reservations").
		Where("user_id = ? AND status != ?", userID, "Dibatalkan").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TotalRevenue).Error
	if err != nil {
		return stats, err
	}

	err = r.getDB().Table("reservations").
		Where("user_id = ? AND status != ?", userID, "Dibatalkan").
		Count(&stats.SoldReservationsCount).Error

	return stats, err
}

func (r *repository) GetTotalExpenses(userID uint) (float64, error) {
	var total float64
	err := r.getDB().Table("expenses").
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *repository) GetExpensesByCategory(userID uint) ([]CategoryExpenseBreakdown, error) {
	var breakdown []CategoryExpenseBreakdown
	err := r.getDB().Table("expenses").
		Select("category, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ?", userID).
		Group("category").
		Order("total desc").
		Scan(&breakdown).Error
	return breakdown, err
}
