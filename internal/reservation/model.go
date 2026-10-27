package reservation

import (
	"fmt"
	"math/rand"
	"time"

	"hotel-app/internal/auth"
	"hotel-app/internal/room"
)

// Daftar konstanta status reservasi dan status pembayaran
const (
	ReservationStatusCheckIn   = "Check-in"
	ReservationStatusCheckOut  = "Check-out"
	ReservationStatusCancelled = "Dibatalkan"

	PaymentStatusPaid    = "Lunas"
	PaymentStatusDeposit = "DP"
	PaymentStatusUnpaid  = "Belum Bayar"

	PaymentMethodCash         = "Tunai"
	PaymentMethodBankTransfer = "Transfer Bank"
	PaymentMethodCard         = "Kartu Debit/Kredit"
	PaymentMethodQRIS         = "QRIS"

	ReminderStatusPending = "Belum Dikirim"
	ReminderStatusSent    = "Terkirim"
	ReminderStatusFailed  = "Gagal"
)

// PaymentMethods memuat daftar seluruh metode pembayaran yang didukung.
var PaymentMethods = []string{
	PaymentMethodCash,
	PaymentMethodBankTransfer,
	PaymentMethodCard,
	PaymentMethodQRIS,
}

// IsValidPaymentMethod memvalidasi keabsahan opsi metode pembayaran.
func IsValidPaymentMethod(method string) bool {
	for _, m := range PaymentMethods {
		if m == method {
			return true
		}
	}
	return false
}

// Reservation merepresentasikan transaksi reservasi/menginap tamu hotel.
type Reservation struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"index;not null" json:"user_id"`
	RoomID          uint       `gorm:"index;not null" json:"room_id"`
	InvoiceNumber   string     `gorm:"size:50;not null;index" json:"invoice_number"`
	GuestName       string     `gorm:"size:100;not null" json:"guest_name"`
	GuestIDCard     string     `gorm:"size:50;not null" json:"guest_id_card"` // KTP / Passport
	GuestPhone      string     `gorm:"size:30" json:"guest_phone"`
	GuestEmail      string     `gorm:"size:150" json:"guest_email"` // Email customer untuk pengingat
	CheckInDate     time.Time  `gorm:"not null" json:"check_in_date"`
	CheckOutDate    time.Time  `gorm:"not null" json:"check_out_date"`
	Nights          int        `gorm:"not null" json:"nights"`
	RoomRate        float64    `gorm:"not null" json:"room_rate"` // Tarif per malam saat reservasi dibuat
	TotalAmount     float64    `gorm:"not null" json:"total_amount"`
	DownPayment     float64    `gorm:"default:0" json:"down_payment"`
	RemainingAmount float64    `gorm:"default:0" json:"remaining_amount"`
	PaymentMethod   string     `gorm:"size:50;not null;default:'Tunai'" json:"payment_method"` // Tunai / Transfer Bank / Kartu Debit/Kredit / QRIS
	PaymentStatus   string     `gorm:"size:30;not null" json:"payment_status"`                 // Lunas / DP / Belum Bayar
	Status          string     `gorm:"size:30;not null" json:"status"`                         // Check-in / Check-out / Dibatalkan
	Notes           string     `gorm:"size:255" json:"notes"`
	ReminderSentAt  *time.Time `json:"reminder_sent_at"`                                       // Waktu notifikasi reminder dikirim
	ReminderStatus  string     `gorm:"size:30;default:'Belum Dikirim'" json:"reminder_status"` // Belum Dikirim / Terkirim / Gagal
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Associations
	User auth.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Room room.Room `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"room"`
}

// TableName menentukan nama tabel database untuk Reservation.
func (Reservation) TableName() string {
	return "reservations"
}

// GenerateInvoiceNumber membuat nomor nota/invoice acak dan unik berformat INV-YYYYMMDD-XXXX
func GenerateInvoiceNumber() string {
	now := time.Now()
	randomNum := rand.Intn(9000) + 1000
	return fmt.Sprintf("INV-%s-%d", now.Format("20060102"), randomNum)
}

// ReservationInput mendefinisikan validasi payload untuk transaksi check-in baru.
type ReservationInput struct {
	RoomID        uint    `json:"room_id" binding:"required"`
	GuestName     string  `json:"guest_name" binding:"required"`
	GuestIDCard   string  `json:"guest_id_card" binding:"required"`
	GuestPhone    string  `json:"guest_phone"`
	GuestEmail    string  `json:"guest_email"`
	CheckInDate   string  `json:"check_in_date" binding:"required"`  // Format: YYYY-MM-DD
	CheckOutDate  string  `json:"check_out_date" binding:"required"` // Format: YYYY-MM-DD
	DownPayment   float64 `json:"down_payment" binding:"gte=0"`
	PaymentMethod string  `json:"payment_method"` // Tunai / Transfer Bank / Kartu Debit/Kredit / QRIS
	Notes         string  `json:"notes"`
}
