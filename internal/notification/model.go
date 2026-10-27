package notification

import "time"

// ReminderPayload memuat seluruh data yang dibutuhkan untuk menyusun pesan reminder checkout.
type ReminderPayload struct {
	ReservationID   uint      `json:"reservation_id"`
	InvoiceNumber   string    `json:"invoice_number"`
	GuestName       string    `json:"guest_name"`
	GuestPhone      string    `json:"guest_phone"`
	GuestEmail      string    `json:"guest_email"`
	RoomNumber      string    `json:"room_number"`
	RoomType        string    `json:"room_type"`
	CheckOutTime    time.Time `json:"check_out_time"`
	RemainingAmount float64   `json:"remaining_amount"`
	HotelName       string    `json:"hotel_name"`
}

// ReminderResult memuat status pengiriman reminder via WhatsApp dan Email.
type ReminderResult struct {
	WhatsAppSent bool   `json:"whatsapp_sent"`
	EmailSent    bool   `json:"email_sent"`
	WhatsAppURL  string `json:"whatsapp_url,omitempty"` // Tautan langsung wa.me untuk staf hotel
	Message      string `json:"message"`
	Details      string `json:"details"`
}
