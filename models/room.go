package models

import (
	"time"
)

// Daftar konstanta status kamar
const (
	RoomStatusAvailable   = "Tersedia"
	RoomStatusOccupied    = "Terisi"
	RoomStatusCleaning    = "Sedang Dibersihkan"
	RoomStatusMaintenance = "Dalam Perbaikan"
)

// Room merepresentasikan data kamar hotel yang dimiliki oleh pengguna tertentu.
type Room struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	RoomNumber    string    `gorm:"size:50;not null;index:idx_user_room,unique" json:"room_number"`
	Type          string    `gorm:"size:50;not null" json:"type"` // Misal: Standard, Deluxe, Suite, dsb.
	PricePerNight float64   `gorm:"not null" json:"price_per_night"`
	Status        string    `gorm:"size:30;not null;default:'Tersedia'" json:"status"`
	Description   string    `gorm:"size:255" json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relasi ke User
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// IsValidStatus memvalidasi apakah status yang dimasukkan sesuai dengan standar sistem.
func IsValidRoomStatus(status string) bool {
	switch status {
	case RoomStatusAvailable, RoomStatusOccupied, RoomStatusCleaning, RoomStatusMaintenance:
		return true
	default:
		return false
	}
}
