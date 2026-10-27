package expense

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler menangani request HTTP untuk domain Biaya Operasional (USALI).
type Handler struct {
	service Service
	dbReady func() bool
}

// NewHandler membuat instance baru dari Expense Handler.
func NewHandler(service Service, dbReady func() bool) *Handler {
	return &Handler{
		service: service,
		dbReady: dbReady,
	}
}

// GetExpenseCategories mengembalikan daftar kategori USALI yang valid.
func (h *Handler) GetExpenseCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": h.service.GetCategories(),
	})
}

// GetExpenses mengambil seluruh catatan pengeluaran milik pengguna dengan filter kategori dan rentang tanggal.
func (h *Handler) GetExpenses(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	categoryQuery := c.Query("category")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	resp, err := h.service.GetExpenses(userID, categoryQuery, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengeluaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         resp.Data,
		"total_amount": resp.TotalAmount,
		"count":        resp.Count,
	})
}

// GetExpenseByID mengambil detail 1 pengeluaran milik pengguna.
func (h *Handler) GetExpenseByID(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengeluaran tidak ditemukan atau bukan milik Anda"})
		return
	}

	expense, err := h.service.GetExpenseByID(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrExpenseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengeluaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": expense,
	})
}

// CreateExpense mencatat pengeluaran operasional baru berstandar USALI.
func (h *Handler) CreateExpense(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	var input ExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: pastikan judul, kategori, jumlah, dan tanggal terisi"})
		return
	}

	expense, err := h.service.CreateExpense(userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidCategory) || errors.Is(err, ErrInvalidDateFormat) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan catatan pengeluaran"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengeluaran operasional berhasil dicatat",
		"data":    expense,
	})
}

// UpdateExpense memperbarui catatan pengeluaran dengan validasi kepemilikan user_id.
func (h *Handler) UpdateExpense(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengeluaran tidak ditemukan atau bukan milik Anda"})
		return
	}

	var input ExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	expense, err := h.service.UpdateExpense(uint(id), userID, input)
	if err != nil {
		if errors.Is(err, ErrExpenseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "Kategori") || errors.Is(err, ErrInvalidDateFormat) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data pengeluaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data pengeluaran berhasil diperbarui",
		"data":    expense,
	})
}

// DeleteExpense menghapus catatan pengeluaran.
func (h *Handler) DeleteExpense(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengeluaran tidak ditemukan atau bukan milik Anda"})
		return
	}

	err = h.service.DeleteExpense(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrExpenseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pengeluaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data pengeluaran berhasil dihapus",
	})
}
