package controllers

import (
	"net/http"
	"time"

	"hotel-app/config"
	"hotel-app/models"

	"github.com/gin-gonic/gin"
)

type ExpenseInput struct {
	Category    string  `json:"category" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ExpenseDate string  `json:"expense_date" binding:"required"` // Format: YYYY-MM-DD
	ReceiptNo   string  `json:"receipt_no"`
	Notes       string  `json:"notes"`
}

// GetExpenseCategories mengembalikan daftar kategori USALI yang valid.
func GetExpenseCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": models.USALICategories,
	})
}

// GetExpenses mengambil seluruh catatan pengeluaran milik pengguna dengan filter kategori dan rentang tanggal.
func GetExpenses(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	categoryQuery := c.Query("category")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := config.DB.Where("user_id = ?", userID)

	if categoryQuery != "" {
		query = query.Where("category = ?", categoryQuery)
	}

	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("expense_date >= ?", t)
		}
	}

	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// Tambahkan 1 hari untuk mencakup seluruh hari end_date
			query = query.Where("expense_date <= ?", t.Add(24*time.Hour))
		}
	}

	var expenses []models.Expense
	if err := query.Order("expense_date desc").Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengeluaran"})
		return
	}

	// Hitung total pengeluaran untuk rekapitulasi cepat
	var totalAmount float64
	for _, exp := range expenses {
		totalAmount += exp.Amount
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         expenses,
		"total_amount": totalAmount,
		"count":        len(expenses),
	})
}

// GetExpenseByID mengambil detail 1 pengeluaran milik pengguna.
func GetExpenseByID(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var expense models.Expense
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengeluaran tidak ditemukan atau bukan milik Anda"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": expense,
	})
}

// CreateExpense mencatat pengeluaran operasional baru berstandar USALI.
func CreateExpense(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	var input ExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: pastikan judul, kategori, jumlah, dan tanggal terisi"})
		return
	}

	if !models.IsValidUSALICategory(input.Category) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori tidak valid. Pilih salah satu kategori standar USALI yang tersedia",
		})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}

	expense := models.Expense{
		UserID:      userID,
		Category:    input.Category,
		Title:       input.Title,
		Amount:      input.Amount,
		ExpenseDate: parsedDate,
		ReceiptNo:   input.ReceiptNo,
		Notes:       input.Notes,
	}

	if err := config.DB.Create(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan catatan pengeluaran"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengeluaran operasional berhasil dicatat",
		"data":    expense,
	})
}

// UpdateExpense memperbarui catatan pengeluaran dengan validasi kepemilikan user_id.
func UpdateExpense(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var expense models.Expense
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengeluaran tidak ditemukan atau bukan milik Anda"})
		return
	}

	var input ExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	if !models.IsValidUSALICategory(input.Category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kategori standar USALI tidak valid"})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}

	expense.Category = input.Category
	expense.Title = input.Title
	expense.Amount = input.Amount
	expense.ExpenseDate = parsedDate
	expense.ReceiptNo = input.ReceiptNo
	expense.Notes = input.Notes

	if err := config.DB.Save(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data pengeluaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data pengeluaran berhasil diperbarui",
		"data":    expense,
	})
}

// DeleteExpense menghapus catatan pengeluaran.
func DeleteExpense(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var expense models.Expense
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengeluaran tidak ditemukan atau bukan milik Anda"})
		return
	}

	if err := config.DB.Delete(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pengeluaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data pengeluaran berhasil dihapus",
	})
}
