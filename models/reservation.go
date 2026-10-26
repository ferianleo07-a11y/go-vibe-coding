package models

import (
	"time"
)

const (
	ReservationStatusCheckIn   = "Check-in"
	ReservationStatusCheckOut  = "Check-out"
	ReservationStatusCancelled = "Dibatalkan"

	PaymentStatusPaid    = "Lunas"
	PaymentStatusDeposit = "DP"
	PaymentStatusUnpaid  = "Belum Bayar"
)

// Reservation merepresentasikan transaksi reservasi/menginap tamu hotel.
type Reservation struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index;not null" json:"user_id"`
	RoomID          uint      `gorm:"index;not null" json:"room_id"`
	InvoiceNumber   string    `gorm:"size:50;not null;index" json:"invoice_number"`
	GuestName       string    `gorm:"size:100;not null" json:"guest_name"`
	GuestIDCard     string    `gorm:"size:50;not null" json:"guest_id_card"` // KTP / Passport
	GuestPhone      string    `gorm:"size:30" json:"guest_phone"`
	CheckInDate     time.Time `gorm:"not null" json:"check_in_date"`
	CheckOutDate    time.Time `gorm:"not null" json:"check_out_date"`
	Nights          int       `gorm:"not null" json:"nights"`
	RoomRate        float64   `gorm:"not null" json:"room_rate"` // Tarif per malam saat reservasi dibuat
	TotalAmount     float64   `gorm:"not null" json:"total_amount"`
	DownPayment     float64   `gorm:"default:0" json:"down_payment"`
	RemainingAmount float64   `gorm:"default:0" json:"remaining_amount"`
	PaymentStatus   string    `gorm:"size:30;not null" json:"payment_status"` // Lunas / DP / Belum Bayar
	Status          string    `gorm:"size:30;not null" json:"status"`         // Check-in / Check-out / Dibatalkan
	Notes           string    `gorm:"size:255" json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Associations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Room Room `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"room"`
}
