package expense

import (
	"time"

	"hotel-app/internal/auth"
)

// Kategori Biaya Operasional Standar USALI (Uniform System of Accounts for the Lodging Industry)
const (
	ExpenseCategoryUtilities       = "Utilities"                // Listrik, Air, Gas, Internet
	ExpenseCategorySupplies        = "Operating Supplies"       // Amenities kamar, Linen, Pembersih
	ExpenseCategoryMaintenance     = "Repairs & Maintenance"    // Servis AC, Plumbing, Cat, Mesin
	ExpenseCategorySalaries        = "Salaries & Wages"         // Gaji staf, Tunjangan, Lembur
	ExpenseCategoryMarketing       = "Marketing & Sales"        // Promosi, Iklan, Komisi OTA
	ExpenseCategoryAdministrative  = "Administrative & General" // ATK, Software, Legal, Asuransi
	ExpenseCategoryFoodAndBeverage = "Food & Beverage"          // Bahan makanan & minuman hotel
	ExpenseCategoryOther           = "Other Expenses"           // Lain-lain
)

// USALICategories mengembalikan daftar semua kategori standar USALI yang didukung.
var USALICategories = []string{
	ExpenseCategoryUtilities,
	ExpenseCategorySupplies,
	ExpenseCategoryMaintenance,
	ExpenseCategorySalaries,
	ExpenseCategoryMarketing,
	ExpenseCategoryAdministrative,
	ExpenseCategoryFoodAndBeverage,
	ExpenseCategoryOther,
}

// Expense merepresentasikan catatan pengeluaran biaya operasional hotel.
type Expense struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	Category    string    `gorm:"size:80;not null" json:"category"`
	Title       string    `gorm:"size:150;not null" json:"title"`
	Amount      float64   `gorm:"not null" json:"amount"`
	ExpenseDate time.Time `gorm:"not null" json:"expense_date"`
	ReceiptNo   string    `gorm:"size:80" json:"receipt_no"` // Nomor Kuitansi / Bukti Pembayaran
	Notes       string    `gorm:"size:255" json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Associations
	User auth.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName menentukan nama tabel database untuk Expense.
func (Expense) TableName() string {
	return "expenses"
}

// IsValidUSALICategory memeriksa apakah kategori yang diinput ada dalam standar USALI.
func IsValidUSALICategory(category string) bool {
	for _, c := range USALICategories {
		if c == category {
			return true
		}
	}
	return false
}

// ExpenseInput mendefinisikan validasi payload pencatatan biaya operasional.
type ExpenseInput struct {
	Category    string  `json:"category" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ExpenseDate string  `json:"expense_date" binding:"required"` // Format: YYYY-MM-DD
	ReceiptNo   string  `json:"receipt_no"`
	Notes       string  `json:"notes"`
}

// ExpensesListResponse mendefinisikan format respons rekapitulasi daftar pengeluaran.
type ExpensesListResponse struct {
	Data        []Expense `json:"data"`
	TotalAmount float64   `json:"total_amount"`
	Count       int       `json:"count"`
}
