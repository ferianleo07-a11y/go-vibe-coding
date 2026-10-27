package expense

import (
	"time"

	"hotel-app/config"

	"gorm.io/gorm"
)

// Repository mendefinisikan antarmuka akses data untuk entitas Pengeluaran.
type Repository interface {
	FindAllByUserID(userID uint, category string, startDate, endDate *time.Time) ([]Expense, error)
	FindByIDAndUserID(id uint, userID uint) (*Expense, error)
	Create(expense *Expense) error
	Update(expense *Expense) error
	Delete(expense *Expense) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository membuat instance baru dari Expense Repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return config.DB
}

func (r *repository) FindAllByUserID(userID uint, category string, startDate, endDate *time.Time) ([]Expense, error) {
	query := r.getDB().Where("user_id = ?", userID)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if startDate != nil {
		query = query.Where("expense_date >= ?", *startDate)
	}

	if endDate != nil {
		// Tambahkan 1 hari untuk mencakup seluruh hari endDate
		query = query.Where("expense_date <= ?", endDate.Add(24*time.Hour))
	}

	var expenses []Expense
	if err := query.Order("expense_date desc").Find(&expenses).Error; err != nil {
		return nil, err
	}

	return expenses, nil
}

func (r *repository) FindByIDAndUserID(id uint, userID uint) (*Expense, error) {
	var expense Expense
	if err := r.getDB().Where("id = ? AND user_id = ?", id, userID).First(&expense).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *repository) Create(expense *Expense) error {
	return r.getDB().Create(expense).Error
}

func (r *repository) Update(expense *Expense) error {
	return r.getDB().Save(expense).Error
}

func (r *repository) Delete(expense *Expense) error {
	return r.getDB().Delete(expense).Error
}
