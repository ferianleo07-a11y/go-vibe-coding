package reservation

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"hotel-app/internal/notification"
	"hotel-app/internal/room"
)

var (
	ErrReservationNotFound    = errors.New("Data reservasi tidak ditemukan atau bukan milik Anda")
	ErrInvalidDateFormat      = errors.New("Format tanggal check-in/check-out harus YYYY-MM-DD")
	ErrInvalidDateRange       = errors.New("Tanggal check-out harus setelah tanggal check-in")
	ErrRoomNotFound           = errors.New("Kamar tidak ditemukan atau bukan milik hotel Anda")
	ErrOnlyCheckInCanCheckout = errors.New("Hanya transaksi berstatus 'Check-in' yang dapat di-check-out")
	ErrOnlyCheckInCanCancel   = errors.New("Hanya reservasi aktif yang dapat dibatalkan")
	ErrInvalidPaymentMethod   = errors.New("Metode pembayaran tidak valid. Pilihan: Tunai, Transfer Bank, Kartu Debit/Kredit, QRIS")
)

// Service mendefinisikan antarmuka logika bisnis untuk entitas Reservasi.
type Service interface {
	GetReservations(userID uint, status string) ([]Reservation, error)
	GetReservationByID(id uint, userID uint) (*Reservation, error)
	CreateReservation(userID uint, input ReservationInput) (*Reservation, error)
	CheckoutReservation(id uint, userID uint) (*Reservation, error)
	CancelReservation(id uint, userID uint) (*Reservation, error)
	SendReminder(id uint, userID uint) (*notification.ReminderResult, *Reservation, error)
	ProcessUpcomingCheckoutReminders() (int, error)
	GetPaymentMethods() []string
}

type service struct {
	repo          Repository
	notifyService notification.Service
}

// NewService membuat instance baru dari Reservation Service.
func NewService(repo Repository, notifyService notification.Service) Service {
	if notifyService == nil {
		notifyService = notification.NewService()
	}
	return &service{
		repo:          repo,
		notifyService: notifyService,
	}
}

func (s *service) GetPaymentMethods() []string {
	return PaymentMethods
}

func (s *service) GetReservations(userID uint, status string) ([]Reservation, error) {
	return s.repo.FindAllByUserID(userID, status)
}

func (s *service) GetReservationByID(id uint, userID uint) (*Reservation, error) {
	res, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, ErrReservationNotFound
	}
	return res, nil
}

func (s *service) CreateReservation(userID uint, input ReservationInput) (*Reservation, error) {
	// Parsing tanggal check-in & check-out
	checkInTime, err := time.Parse("2006-01-02", input.CheckInDate)
	if err != nil {
		return nil, errors.New("Format tanggal check-in harus YYYY-MM-DD")
	}

	checkOutTime, err := time.Parse("2006-01-02", input.CheckOutDate)
	if err != nil {
		return nil, errors.New("Format tanggal check-out harus YYYY-MM-DD")
	}

	if !checkOutTime.After(checkInTime) {
		return nil, ErrInvalidDateRange
	}

	// Standarisasi jam check-out hotel pukul 12:00:00 siang
	checkOutTime = time.Date(checkOutTime.Year(), checkOutTime.Month(), checkOutTime.Day(), 12, 0, 0, 0, time.Local)
	checkInTime = time.Date(checkInTime.Year(), checkInTime.Month(), checkInTime.Day(), 14, 0, 0, 0, time.Local)

	// Hitung durasi malam
	durationHours := checkOutTime.Sub(checkInTime).Hours()
	nights := int(durationHours / 24)
	if nights < 1 {
		nights = 1
	}

	// Validasi & default metode pembayaran
	paymentMethod := strings.TrimSpace(input.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = PaymentMethodCash
	} else if !IsValidPaymentMethod(paymentMethod) {
		return nil, ErrInvalidPaymentMethod
	}

	// Ambil data kamar & pastikan milik pengguna ini
	roomData, err := s.repo.FindRoomByIDAndUserID(input.RoomID, userID)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	if roomData.Status != room.RoomStatusAvailable {
		return nil, fmt.Errorf("Kamar nomor %s saat ini sedang '%s'. Hanya kamar 'Tersedia' yang dapat di-check-in", roomData.RoomNumber, roomData.Status)
	}

	// Hitung total biaya & status pembayaran
	totalAmount := float64(nights) * roomData.PricePerNight
	remainingAmount := totalAmount - input.DownPayment
	if remainingAmount < 0 {
		remainingAmount = 0
	}

	paymentStatus := PaymentStatusUnpaid
	if input.DownPayment >= totalAmount {
		paymentStatus = PaymentStatusPaid
	} else if input.DownPayment > 0 {
		paymentStatus = PaymentStatusDeposit
	}

	reservation := Reservation{
		UserID:          userID,
		RoomID:          roomData.ID,
		InvoiceNumber:   GenerateInvoiceNumber(),
		GuestName:       input.GuestName,
		GuestIDCard:     input.GuestIDCard,
		GuestPhone:      input.GuestPhone,
		GuestEmail:      input.GuestEmail,
		CheckInDate:     checkInTime,
		CheckOutDate:    checkOutTime,
		Nights:          nights,
		RoomRate:        roomData.PricePerNight,
		TotalAmount:     totalAmount,
		DownPayment:     input.DownPayment,
		RemainingAmount: remainingAmount,
		PaymentMethod:   paymentMethod,
		PaymentStatus:   paymentStatus,
		Status:          ReservationStatusCheckIn,
		ReminderStatus:  ReminderStatusPending,
		Notes:           input.Notes,
	}

	if err := s.repo.CreateReservationTx(&reservation, roomData); err != nil {
		return nil, err
	}

	// Ambil ulang reservasi lengkap dengan relasi Room
	created, err := s.repo.FindByIDAndUserID(reservation.ID, userID)
	if err != nil {
		reservation.Room = *roomData
		return &reservation, nil
	}

	return created, nil
}

func (s *service) CheckoutReservation(id uint, userID uint) (*Reservation, error) {
	reservation, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, ErrReservationNotFound
	}

	if reservation.Status != ReservationStatusCheckIn {
		return nil, fmt.Errorf("Hanya transaksi berstatus 'Check-in' yang dapat di-check-out. Status saat ini: %s", reservation.Status)
	}

	if err := s.repo.CheckoutReservationTx(reservation); err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *service) CancelReservation(id uint, userID uint) (*Reservation, error) {
	reservation, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, ErrReservationNotFound
	}

	if reservation.Status != ReservationStatusCheckIn {
		return nil, ErrOnlyCheckInCanCancel
	}

	if err := s.repo.CancelReservationTx(reservation); err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *service) SendReminder(id uint, userID uint) (*notification.ReminderResult, *Reservation, error) {
	reservation, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, nil, ErrReservationNotFound
	}

	if reservation.Status != ReservationStatusCheckIn {
		return nil, nil, errors.New("Pengingat hanya dapat dikirim untuk tamu yang berstatus aktif 'Check-in'")
	}

	hotelName := "Hotel Management"
	if reservation.User.Name != "" {
		hotelName = reservation.User.Name
	}

	payload := notification.ReminderPayload{
		ReservationID:   reservation.ID,
		InvoiceNumber:   reservation.InvoiceNumber,
		GuestName:       reservation.GuestName,
		GuestPhone:      reservation.GuestPhone,
		GuestEmail:      reservation.GuestEmail,
		RoomNumber:      reservation.Room.RoomNumber,
		RoomType:        reservation.Room.Type,
		CheckOutTime:    reservation.CheckOutDate,
		RemainingAmount: reservation.RemainingAmount,
		HotelName:       hotelName,
	}

	result, err := s.notifyService.SendCheckoutReminder(payload)
	if err != nil {
		now := time.Now()
		_ = s.repo.UpdateReminderStatus(reservation.ID, ReminderStatusFailed, &now)
		return nil, nil, fmt.Errorf("Gagal mengirim pengingat: %w", err)
	}

	now := time.Now()
	_ = s.repo.UpdateReminderStatus(reservation.ID, ReminderStatusSent, &now)
	reservation.ReminderStatus = ReminderStatusSent
	reservation.ReminderSentAt = &now

	return result, reservation, nil
}

// ProcessUpcomingCheckoutReminders memproses otomatis pengiriman reminder 15 menit sebelum checkout.
func (s *service) ProcessUpcomingCheckoutReminders() (int, error) {
	// Jendela waktu reminder: saat ini sampai 15 menit ke depan
	windowEnd := time.Now().Add(15 * time.Minute)

	reservations, err := s.repo.FindUpcomingCheckoutsForReminder(windowEnd)
	if err != nil {
		return 0, err
	}

	sentCount := 0
	for _, res := range reservations {
		hotelName := "Hotel Management"
		if res.User.Name != "" {
			hotelName = res.User.Name
		}

		payload := notification.ReminderPayload{
			ReservationID:   res.ID,
			InvoiceNumber:   res.InvoiceNumber,
			GuestName:       res.GuestName,
			GuestPhone:      res.GuestPhone,
			GuestEmail:      res.GuestEmail,
			RoomNumber:      res.Room.RoomNumber,
			RoomType:        res.Room.Type,
			CheckOutTime:    res.CheckOutDate,
			RemainingAmount: res.RemainingAmount,
			HotelName:       hotelName,
		}

		_, err := s.notifyService.SendCheckoutReminder(payload)
		now := time.Now()
		if err != nil {
			log.Printf("[Auto Reminder Worker] Gagal mengirim reminder untuk reservasi #%d (%s): %v\n", res.ID, res.GuestName, err)
			_ = s.repo.UpdateReminderStatus(res.ID, ReminderStatusFailed, &now)
		} else {
			log.Printf("[Auto Reminder Worker] Berhasil mengirim reminder checkout 15 menit untuk tamu %s (Kamar %s)\n", res.GuestName, res.Room.RoomNumber)
			_ = s.repo.UpdateReminderStatus(res.ID, ReminderStatusSent, &now)
			sentCount++
		}
	}

	return sentCount, nil
}
