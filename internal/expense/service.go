package expense

import (
	"errors"
	"time"
)

var (
	ErrExpenseNotFound   = errors.New("Data pengeluaran tidak ditemukan atau bukan milik Anda")
	ErrInvalidCategory    = errors.New("Kategori tidak valid. Pilih salah satu kategori standar USALI yang tersedia")
	ErrInvalidDateFormat  = errors.New("Format tanggal harus YYYY-MM-DD")
)

// Service mendefinisikan antarmuka logika bisnis untuk entitas Pengeluaran.
type Service interface {
	GetCategories() []string
	GetExpenses(userID uint, category string, startDateStr, endDateStr string) (*ExpensesListResponse, error)
	GetExpenseByID(id uint, userID uint) (*Expense, error)
	CreateExpense(userID uint, input ExpenseInput) (*Expense, error)
	UpdateExpense(id uint, userID uint, input ExpenseInput) (*Expense, error)
	DeleteExpense(id uint, userID uint) error
}

type service struct {
	repo Repository
}

// NewService membuat instance baru dari Expense Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetCategories() []string {
	return USALICategories
}

func (s *service) GetExpenses(userID uint, category string, startDateStr, endDateStr string) (*ExpensesListResponse, error) {
	var startPtr, endPtr *time.Time

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startPtr = &t
		}
	}

	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endPtr = &t
		}
	}

	expenses, err := s.repo.FindAllByUserID(userID, category, startPtr, endPtr)
	if err != nil {
		return nil, err
	}

	var totalAmount float64
	for _, exp := range expenses {
		totalAmount += exp.Amount
	}

	return &ExpensesListResponse{
		Data:        expenses,
		TotalAmount: totalAmount,
		Count:       len(expenses),
	}, nil
}

func (s *service) GetExpenseByID(id uint, userID uint) (*Expense, error) {
	exp, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, ErrExpenseNotFound
	}
	return exp, nil
}

func (s *service) CreateExpense(userID uint, input ExpenseInput) (*Expense, error) {
	if !IsValidUSALICategory(input.Category) {
		return nil, ErrInvalidCategory
	}

	parsedDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	expense := Expense{
		UserID:      userID,
		Category:    input.Category,
		Title:       input.Title,
		Amount:      input.Amount,
		ExpenseDate: parsedDate,
		ReceiptNo:   input.ReceiptNo,
		Notes:       input.Notes,
	}

	if err := s.repo.Create(&expense); err != nil {
		return nil, err
	}

	return &expense, nil
}

func (s *service) UpdateExpense(id uint, userID uint, input ExpenseInput) (*Expense, error) {
	expense, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, ErrExpenseNotFound
	}

	if !IsValidUSALICategory(input.Category) {
		return nil, errors.New("Kategori standar USALI tidak valid")
	}

	parsedDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	expense.Category = input.Category
	expense.Title = input.Title
	expense.Amount = input.Amount
	expense.ExpenseDate = parsedDate
	expense.ReceiptNo = input.ReceiptNo
	expense.Notes = input.Notes

	if err := s.repo.Update(expense); err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *service) DeleteExpense(id uint, userID uint) error {
	expense, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return ErrExpenseNotFound
	}

	return s.repo.Delete(expense)
}
