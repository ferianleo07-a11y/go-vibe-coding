package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Service mendefinisikan antarmuka pengiriman notifikasi pengingat tamu.
type Service interface {
	SendCheckoutReminder(payload ReminderPayload) (*ReminderResult, error)
	FormatWhatsAppMessage(payload ReminderPayload) string
	FormatEmailBody(payload ReminderPayload) (string, string)
}

type service struct{}

// NewService membuat instance baru dari Notification Service.
func NewService() Service {
	return &service{}
}

// FormatWhatsAppMessage menyusun template pesan WhatsApp reminder yang ramah & informatif.
func (s *service) FormatWhatsAppMessage(payload ReminderPayload) string {
	hotelName := payload.HotelName
	if hotelName == "" {
		hotelName = "Hotel Management"
	}

	checkoutStr := payload.CheckOutTime.Format("15:04 WIB (02-Jan-2006)")

	var paymentInfo string
	if payload.RemainingAmount > 0 {
		paymentInfo = fmt.Sprintf("\n💳 *Sisa Tagihan:* Rp %s (mohon diselesaikan saat check-out)", formatCurrency(payload.RemainingAmount))
	} else {
		paymentInfo = "\n💳 *Status Pembayaran:* LUNAS"
	}

	return fmt.Sprintf(
		"🔔 *PENGINGAT CHECK-OUT (15 MENIT LAGI)*\n\n"+
			"Halo Bapak/Ibu *%s*,\n\n"+
			"Waktu check-out Anda untuk *Kamar %s* (%s) dijadwalkan pada pukul *%s*.\n\n"+
			"📋 *No. Invoice:* %s%s\n\n"+
			"Mohon pastikan seluruh barang bawaan pribadi Anda tidak ada yang tertinggal, dan silakan menuju meja resepsionis untuk pengembalian kunci kamar.\n\n"+
			"Terima kasih telah menginap di *%s*. Semoga perjalanan Anda menyenangkan! 🙏",
		payload.GuestName,
		payload.RoomNumber,
		payload.RoomType,
		checkoutStr,
		payload.InvoiceNumber,
		paymentInfo,
		hotelName,
	)
}

// FormatEmailBody menyusun konten email reminder dalam format Plain Text dan HTML.
func (s *service) FormatEmailBody(payload ReminderPayload) (string, string) {
	hotelName := payload.HotelName
	if hotelName == "" {
		hotelName = "Hotel Management"
	}

	checkoutStr := payload.CheckOutTime.Format("15:04 WIB (02-Jan-2006)")

	var paymentNotice string
	if payload.RemainingAmount > 0 {
		paymentNotice = fmt.Sprintf("Sisa Tagihan: Rp %s (dapat diselesaikan di resepsionis)", formatCurrency(payload.RemainingAmount))
	} else {
		paymentNotice = "Status Tagihan: LUNAS"
	}

	plainText := fmt.Sprintf(
		"Halo %s,\n\n"+
			"Ini adalah pengingat bahwa waktu check-out kamar Anda di %s (Kamar %s - %s) dijadwalkan pada pukul %s (15 menit lagi).\n\n"+
			"Invoice: %s\n"+
			"%s\n\n"+
			"Mohon persiapkan barang bawaan Anda dan silakan menuju ke resepsionis untuk pengembalian kunci kamar.\n\n"+
			"Salam hangat,\n%s",
		payload.GuestName, hotelName, payload.RoomNumber, payload.RoomType, checkoutStr,
		payload.InvoiceNumber, paymentNotice, hotelName,
	)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Pengingat Check-out Hotel</title>
</head>
<body style="font-family: Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b;">
  <div style="max-width: 580px; margin: 0 auto; background: #ffffff; border-radius: 8px; border: 1px solid #e2e8f0; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.05);">
    <div style="background: #1e3a8a; padding: 20px; text-align: center; color: #ffffff;">
      <h2 style="margin: 0; font-size: 20px;">%s</h2>
      <p style="margin: 5px 0 0; font-size: 13px; opacity: 0.9;">Pengingat Check-out 15 Menit</p>
    </div>
    <div style="padding: 24px;">
      <p style="font-size: 16px;">Halo <strong>%s</strong>,</p>
      <p style="font-size: 14px; line-height: 1.6;">
        Waktu check-out untuk kamar Anda dijadwalkan dalam <strong>15 menit ke depan</strong> pada pukul <strong>%s</strong>.
      </p>

      <div style="background: #f1f5f9; border-radius: 6px; padding: 16px; margin: 20px 0; font-size: 14px;">
        <div style="margin-bottom: 8px;"><strong>Nomor Kamar:</strong> %s (%s)</div>
        <div style="margin-bottom: 8px;"><strong>Nomor Invoice:</strong> %s</div>
        <div style="margin-bottom: 0;"><strong>Status Pembayaran:</strong> <span style="color: #047857; font-weight: bold;">%s</span></div>
      </div>

      <p style="font-size: 14px; line-height: 1.6;">
        Mohon periksa kembali barang bawaan pribadi Anda agar tidak ada yang tertinggal, dan serahkan kunci kamar kepada staf resepsionis kami.
      </p>
      
      <p style="font-size: 14px; margin-top: 24px;">
        Terima kasih telah mempercayakan penginapan Anda kepada kami!<br>
        <strong>%s</strong>
      </p>
    </div>
  </div>
</body>
</html>`,
		hotelName,
		payload.GuestName,
		checkoutStr,
		payload.RoomNumber,
		payload.RoomType,
		payload.InvoiceNumber,
		paymentNotice,
		hotelName,
	)

	return plainText, htmlBody
}

func (s *service) SendCheckoutReminder(payload ReminderPayload) (*ReminderResult, error) {
	waMsg := s.FormatWhatsAppMessage(payload)
	cleanPhone := cleanPhoneNumber(payload.GuestPhone)

	// Buat tautan wa.me standar
	waURL := ""
	if cleanPhone != "" {
		waURL = fmt.Sprintf("https://wa.me/%s?text=%s", cleanPhone, url.QueryEscape(waMsg))
	}

	result := &ReminderResult{
		WhatsAppURL: waURL,
	}

	// 1. Pengiriman WhatsApp
	waSent := false
	waGatewayURL := os.Getenv("WHATSAPP_API_URL")
	if waGatewayURL != "" && cleanPhone != "" {
		// Kirim via WhatsApp Gateway Provider (misal Fonnte, Waba, dll)
		payloadMap := map[string]string{
			"target":  cleanPhone,
			"message": waMsg,
		}
		jsonBytes, _ := json.Marshal(payloadMap)
		req, err := http.NewRequest("POST", waGatewayURL, bytes.NewBuffer(jsonBytes))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			if token := os.Getenv("WHATSAPP_API_TOKEN"); token != "" {
				req.Header.Set("Authorization", token)
			}
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				waSent = true
				resp.Body.Close()
			}
		}
	} else if cleanPhone != "" {
		// Mode Simulasi / Direct WhatsApp Web Link
		log.Printf("[Reminder WA] Pengingat WhatsApp disiapkan untuk %s (%s). Tautan: %s\n", payload.GuestName, cleanPhone, waURL)
		waSent = true
	}
	result.WhatsAppSent = waSent

	// 2. Pengiriman Email
	emailSent := false
	if payload.GuestEmail != "" {
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := os.Getenv("SMTP_PORT")
		smtpUser := os.Getenv("SMTP_USER")
		smtpPass := os.Getenv("SMTP_PASSWORD")
		smtpFrom := os.Getenv("SMTP_FROM")

		if smtpHost != "" && smtpPort != "" && smtpUser != "" && smtpPass != "" {
			if smtpFrom == "" {
				smtpFrom = smtpUser
			}

			hotelName := payload.HotelName
			if hotelName == "" {
				hotelName = "Hotel Management"
			}

			_, htmlContent := s.FormatEmailBody(payload)
			subject := fmt.Sprintf("[Pengingat Check-out] Kamar %s - %s", payload.RoomNumber, hotelName)

			mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
			msg := []byte(fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s", hotelName, smtpFrom, payload.GuestEmail, subject, mime, htmlContent))

			addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
			auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

			err := smtp.SendMail(addr, auth, smtpFrom, []string{payload.GuestEmail}, msg)
			if err != nil {
				log.Printf("[Reminder Email] Gagal mengirim email ke %s: %v\n", payload.GuestEmail, err)
			} else {
				emailSent = true
				log.Printf("[Reminder Email] Berhasil mengirim email reminder ke %s\n", payload.GuestEmail)
			}
		} else {
			// Mode Simulasi / Development Logging
			log.Printf("[Reminder Email (Simulasi)] Reminder dikirimkan ke email tamu %s (%s)\n", payload.GuestName, payload.GuestEmail)
			emailSent = true
		}
	}
	result.EmailSent = emailSent

	var detailParts []string
	if result.WhatsAppSent {
		detailParts = append(detailParts, "WhatsApp terkirim")
	}
	if result.EmailSent {
		detailParts = append(detailParts, "Email terkirim")
	}

	result.Message = "Pengingat check-out berhasil diproses"
	result.Details = strings.Join(detailParts, ", ")

	return result, nil
}

// cleanPhoneNumber merapikan format nomor telepon lokal Indonesia ke format internasional (misal 0812 -> 62812).
func cleanPhoneNumber(phone string) string {
	reg := regexp.MustCompile(`[^0-9]`)
	clean := reg.ReplaceAllString(phone, "")
	if strings.HasPrefix(clean, "08") {
		clean = "628" + clean[2:]
	} else if strings.HasPrefix(clean, "+62") {
		clean = clean[1:]
	}
	return clean
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("%.0f", amount)
}
