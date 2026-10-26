package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"hotel-app/config"
	"hotel-app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReservationInput struct {
	RoomID       uint    `json:"room_id" binding:"required"`
	GuestName    string  `json:"guest_name" binding:"required"`
	GuestIDCard  string  `json:"guest_id_card" binding:"required"`
	GuestPhone   string  `json:"guest_phone"`
	CheckInDate  string  `json:"check_in_date" binding:"required"`  // Format: YYYY-MM-DD
	CheckOutDate string  `json:"check_out_date" binding:"required"` // Format: YYYY-MM-DD
	DownPayment  float64 `json:"down_payment" binding:"gte=0"`
	Notes        string  `json:"notes"`
}

// GenerateInvoiceNumber membuat nomor nota/invoice acak dan unik berformat INV-YYYYMMDD-XXXX
func GenerateInvoiceNumber() string {
	now := time.Now()
	randomNum := rand.Intn(9000) + 1000
	return fmt.Sprintf("INV-%s-%d", now.Format("20060102"), randomNum)
}

// GetReservations mengambil seluruh daftar reservasi milik pengguna dengan opsi filter status.
func GetReservations(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	statusQuery := c.Query("status")

	query := config.DB.Where("user_id = ?", userID).Preload("Room")
	if statusQuery != "" {
		query = query.Where("status = ?", statusQuery)
	}

	var reservations []models.Reservation
	if err := query.Order("created_at desc").Find(&reservations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data reservasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reservations,
	})
}

// GetReservationByID mengambil detail reservasi beserta data kamar.
func GetReservationByID(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var reservation models.Reservation
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).Preload("Room").First(&reservation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan atau bukan milik Anda"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reservation,
	})
}

// CreateReservation mencatat check-in baru, kalkulasi otomatis durasi dan tagihan, serta memperbarui status kamar menjadi 'Terisi'.
func CreateReservation(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	var input ReservationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: pastikan data tamu, tanggal, dan kamar terisi dengan benar"})
		return
	}

	// Parsing tanggal check-in & check-out
	checkInTime, err := time.Parse("2006-01-02", input.CheckInDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal check-in harus YYYY-MM-DD"})
		return
	}

	checkOutTime, err := time.Parse("2006-01-02", input.CheckOutDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal check-out harus YYYY-MM-DD"})
		return
	}

	if !checkOutTime.After(checkInTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tanggal check-out harus setelah tanggal check-in"})
		return
	}

	// Hitung durasi malam
	durationHours := checkOutTime.Sub(checkInTime).Hours()
	nights := int(durationHours / 24)
	if nights < 1 {
		nights = 1
	}

	// Ambil data kamar & pastikan milik pengguna ini
	var room models.Room
	if err := config.DB.Where("id = ? AND user_id = ?", input.RoomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kamar tidak ditemukan atau bukan milik hotel Anda"})
		return
	}

	if room.Status != models.RoomStatusAvailable {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Kamar nomor %s saat ini sedang '%s'. Hanya kamar 'Tersedia' yang dapat di-check-in", room.RoomNumber, room.Status),
		})
		return
	}

	// Hitung total biaya & status pembayaran
	totalAmount := float64(nights) * room.PricePerNight
	remainingAmount := totalAmount - input.DownPayment
	if remainingAmount < 0 {
		remainingAmount = 0
	}

	paymentStatus := models.PaymentStatusUnpaid
	if input.DownPayment >= totalAmount {
		paymentStatus = models.PaymentStatusPaid
	} else if input.DownPayment > 0 {
		paymentStatus = models.PaymentStatusDeposit
	}

	// Buat transaksi database menggunakan DB Transaction untuk konsistensi data
	var reservation models.Reservation
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		reservation = models.Reservation{
			UserID:          userID,
			RoomID:          room.ID,
			InvoiceNumber:   GenerateInvoiceNumber(),
			GuestName:       input.GuestName,
			GuestIDCard:     input.GuestIDCard,
			GuestPhone:      input.GuestPhone,
			CheckInDate:     checkInTime,
			CheckOutDate:    checkOutTime,
			Nights:          nights,
			RoomRate:        room.PricePerNight,
			TotalAmount:     totalAmount,
			DownPayment:     input.DownPayment,
			RemainingAmount: remainingAmount,
			PaymentStatus:   paymentStatus,
			Status:          models.ReservationStatusCheckIn,
			Notes:           input.Notes,
		}

		if err := tx.Create(&reservation).Error; err != nil {
			return err
		}

		// Update status kamar menjadi 'Terisi'
		room.Status = models.RoomStatusOccupied
		if err := tx.Save(&room).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses transaksi reservasi"})
		return
	}

	// Load room relation for response
	config.DB.Preload("Room").First(&reservation, reservation.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Proses check-in berhasil dicatat",
		"data":    reservation,
	})
}

// CheckoutReservation memproses check-out tamu, melunasi sisa tagihan, dan mengubah status kamar menjadi 'Sedang Dibersihkan'.
func CheckoutReservation(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var reservation models.Reservation
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).Preload("Room").First(&reservation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
		return
	}

	if reservation.Status != models.ReservationStatusCheckIn {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Hanya transaksi berstatus 'Check-in' yang dapat di-check-out. Status saat ini: %s", reservation.Status),
		})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Update status reservasi
		reservation.Status = models.ReservationStatusCheckOut
		reservation.PaymentStatus = models.PaymentStatusPaid
		reservation.RemainingAmount = 0
		if err := tx.Save(&reservation).Error; err != nil {
			return err
		}

		// Update status kamar menjadi 'Sedang Dibersihkan'
		var room models.Room
		if err := tx.Where("id = ?", reservation.RoomID).First(&room).Error; err == nil {
			room.Status = models.RoomStatusCleaning
			if err := tx.Save(&room).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses check-out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Proses check-out berhasil. Kamar kini berstatus 'Sedang Dibersihkan'",
		"data":    reservation,
	})
}

// CancelReservation membatalkan reservasi dan mengembalikan status kamar menjadi 'Tersedia'.
func CancelReservation(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var reservation models.Reservation
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&reservation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data reservasi tidak ditemukan"})
		return
	}

	if reservation.Status != models.ReservationStatusCheckIn {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Hanya reservasi aktif yang dapat dibatalkan",
		})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		reservation.Status = models.ReservationStatusCancelled
		if err := tx.Save(&reservation).Error; err != nil {
			return err
		}

		// Kembalikan status kamar menjadi 'Tersedia'
		var room models.Room
		if err := tx.Where("id = ?", reservation.RoomID).First(&room).Error; err == nil {
			room.Status = models.RoomStatusAvailable
			if err := tx.Save(&room).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membatalkan reservasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reservasi berhasil dibatalkan dan kamar telah kembali 'Tersedia'",
		"data":    reservation,
	})
}
