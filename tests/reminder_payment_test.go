package tests

import (
	"strings"
	"testing"
	"time"

	"hotel-app/internal/notification"
	"hotel-app/internal/reservation"
)

// TestPaymentMethodValidation menguji keabsahan opsi metode pembayaran pada reservasi
func TestPaymentMethodValidation(t *testing.T) {
	validMethods := []string{
		"Tunai",
		"Transfer Bank",
		"Kartu Debit/Kredit",
		"QRIS",
	}

	for _, method := range validMethods {
		if !reservation.IsValidPaymentMethod(method) {
			t.Errorf("Metode pembayaran valid '%s' seharusnya diterima", method)
		}
	}

	invalidMethods := []string{
		"PayLater",
		"Cryptocurrency",
		"Utang",
		"",
	}

	for _, method := range invalidMethods {
		if reservation.IsValidPaymentMethod(method) {
			t.Errorf("Metode pembayaran tidak valid '%s' seharusnya ditolak", method)
		}
	}
}

// TestNotificationCheckoutReminderFormatting menguji pembentukan pesan reminder WA dan Email
func TestNotificationCheckoutReminderFormatting(t *testing.T) {
	notifyService := notification.NewService()

	checkoutTime := time.Date(2026, 10, 27, 12, 0, 0, 0, time.Local)
	payload := notification.ReminderPayload{
		ReservationID:   1,
		InvoiceNumber:   "INV-20261027-1234",
		GuestName:       "Budi Santoso",
		GuestPhone:      "081234567890",
		GuestEmail:      "budi@example.com",
		RoomNumber:      "101",
		RoomType:        "Deluxe Room",
		CheckOutTime:    checkoutTime,
		RemainingAmount: 250000,
		HotelName:       "Grand Hotel Harmoni",
	}

	// 1. Uji template WhatsApp
	waMsg := notifyService.FormatWhatsAppMessage(payload)
	if !strings.Contains(waMsg, "Budi Santoso") {
		t.Errorf("Pesan WA harus memuat nama tamu, didapatkan: %s", waMsg)
	}
	if !strings.Contains(waMsg, "101") {
		t.Errorf("Pesan WA harus memuat nomor kamar 101, didapatkan: %s", waMsg)
	}
	if !strings.Contains(waMsg, "INV-20261027-1234") {
		t.Errorf("Pesan WA harus memuat nomor invoice, didapatkan: %s", waMsg)
	}
	if !strings.Contains(waMsg, "250000") {
		t.Errorf("Pesan WA harus memuat sisa tagihan, didapatkan: %s", waMsg)
	}

	// 2. Uji template Email
	plainText, htmlBody := notifyService.FormatEmailBody(payload)
	if !strings.Contains(plainText, "Budi Santoso") || !strings.Contains(htmlBody, "Budi Santoso") {
		t.Errorf("Email harus memuat nama tamu")
	}
	if !strings.Contains(htmlBody, "15 menit") {
		t.Errorf("Email HTML harus menegaskan pengingat 15 menit menuju checkout")
	}

	// 3. Uji pengiriman (mode simulasi/development)
	result, err := notifyService.SendCheckoutReminder(payload)
	if err != nil {
		t.Fatalf("SendCheckoutReminder gagal: %v", err)
	}

	if !result.WhatsAppSent {
		t.Errorf("WhatsApp seharusnya berhasil disiapkan/dikirim")
	}
	if !result.EmailSent {
		t.Errorf("Email seharusnya berhasil dikirimkan/disimulasikan")
	}
	if !strings.Contains(result.WhatsAppURL, "wa.me/6281234567890") {
		t.Errorf("Tautan WhatsApp harus diformat ke kode negara 62, didapatkan: %s", result.WhatsAppURL)
	}
}
