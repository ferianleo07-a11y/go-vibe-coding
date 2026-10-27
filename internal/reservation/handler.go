package reservation

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler menangani request HTTP untuk domain Reservasi.
type Handler struct {
	service Service
	dbReady func() bool
}

// NewHandler membuat instance baru dari Reservation Handler.
func NewHandler(service Service, dbReady func() bool) *Handler {
	return &Handler{
		service: service,
		dbReady: dbReady,
	}
}

// GetPaymentMethods mengembalikan daftar opsi metode pembayaran yang valid.
func (h *Handler) GetPaymentMethods(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": h.service.GetPaymentMethods(),
	})
}

// GetReservations mengambil seluruh daftar reservasi milik pengguna dengan opsi filter status.
func (h *Handler) GetReservations(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	statusQuery := c.Query("status")

	reservations, err := h.service.GetReservations(userID, statusQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data reservasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reservations,
	})
}

// GetReservationByID mengambil detail reservasi beserta data kamar.
func (h *Handler) GetReservationByID(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan atau bukan milik Anda"})
		return
	}

	reservation, err := h.service.GetReservationByID(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data reservasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reservation,
	})
}

// CreateReservation mencatat check-in baru, kalkulasi otomatis durasi dan tagihan, serta memperbarui status kamar menjadi 'Terisi'.
func (h *Handler) CreateReservation(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	var input ReservationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: pastikan data tamu, tanggal, dan kamar terisi dengan benar"})
		return
	}

	reservation, err := h.service.CreateReservation(userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidDateRange) || errors.Is(err, ErrRoomNotFound) || errors.Is(err, ErrInvalidPaymentMethod) ||
			strings.HasPrefix(err.Error(), "Format tanggal") || strings.Contains(err.Error(), "Hanya kamar 'Tersedia'") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses transaksi reservasi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Proses check-in berhasil dicatat",
		"data":    reservation,
	})
}

// CheckoutReservation memproses check-out tamu, melunasi sisa tagihan, dan mengubah status kamar menjadi 'Sedang Dibersihkan'.
func (h *Handler) CheckoutReservation(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
		return
	}

	reservation, err := h.service.CheckoutReservation(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
			return
		}
		if strings.Contains(err.Error(), "Hanya transaksi berstatus 'Check-in'") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses check-out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Proses check-out berhasil. Kamar kini berstatus 'Sedang Dibersihkan'",
		"data":    reservation,
	})
}

// CancelReservation membatalkan reservasi dan mengembalikan status kamar menjadi 'Tersedia'.
func (h *Handler) CancelReservation(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
		return
	}

	reservation, err := h.service.CancelReservation(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
			return
		}
		if errors.Is(err, ErrOnlyCheckInCanCancel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membatalkan reservasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reservasi berhasil dibatalkan dan kamar telah kembali 'Tersedia'",
		"data":    reservation,
	})
}

// SendReminder mengirimkan notifikasi pengingat WhatsApp dan Email kepada customer menjelang waktu check-out.
func (h *Handler) SendReminder(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
		return
	}

	result, updatedReservation, err := h.service.SendReminder(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Pengingat check-out 15 menit berhasil dikirimkan via WhatsApp dan Email",
		"whatsapp_sent": result.WhatsAppSent,
		"whatsapp_url":  result.WhatsAppURL,
		"email_sent":    result.EmailSent,
		"details":       result.Details,
		"data":          updatedReservation,
	})
}
