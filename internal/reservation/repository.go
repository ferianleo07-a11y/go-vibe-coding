package reservation

import (
	"time"

	"hotel-app/config"
	"hotel-app/internal/room"

	"gorm.io/gorm"
)

// Repository mendefinisikan antarmuka akses data untuk entitas Reservasi.
type Repository interface {
	FindAllByUserID(userID uint, status string) ([]Reservation, error)
	FindByIDAndUserID(id uint, userID uint) (*Reservation, error)
	FindRoomByIDAndUserID(roomID uint, userID uint) (*room.Room, error)
	CreateReservationTx(reservation *Reservation, roomData *room.Room) error
	CheckoutReservationTx(reservation *Reservation) error
	CancelReservationTx(reservation *Reservation) error
	UpdateReminderStatus(id uint, status string, sentAt *time.Time) error
	FindUpcomingCheckoutsForReminder(windowEnd time.Time) ([]Reservation, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository membuat instance baru dari Reservation Repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return config.DB
}

func (r *repository) FindAllByUserID(userID uint, status string) ([]Reservation, error) {
	query := r.getDB().Where("user_id = ?", userID).Preload("Room")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var reservations []Reservation
	if err := query.Order("created_at desc").Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}

func (r *repository) FindByIDAndUserID(id uint, userID uint) (*Reservation, error) {
	var reservation Reservation
	if err := r.getDB().Where("id = ? AND user_id = ?", id, userID).Preload("Room").First(&reservation).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *repository) FindRoomByIDAndUserID(roomID uint, userID uint) (*room.Room, error) {
	var roomData room.Room
	if err := r.getDB().Where("id = ? AND user_id = ?", roomID, userID).First(&roomData).Error; err != nil {
		return nil, err
	}
	return &roomData, nil
}

func (r *repository) CreateReservationTx(reservation *Reservation, roomData *room.Room) error {
	return r.getDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(reservation).Error; err != nil {
			return err
		}

		roomData.Status = room.RoomStatusOccupied
		if err := tx.Save(roomData).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *repository) CheckoutReservationTx(reservation *Reservation) error {
	return r.getDB().Transaction(func(tx *gorm.DB) error {
		reservation.Status = ReservationStatusCheckOut
		reservation.PaymentStatus = PaymentStatusPaid
		reservation.RemainingAmount = 0
		if err := tx.Save(reservation).Error; err != nil {
			return err
		}

		var roomData room.Room
		if err := tx.Where("id = ?", reservation.RoomID).First(&roomData).Error; err == nil {
			roomData.Status = room.RoomStatusCleaning
			if err := tx.Save(&roomData).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) CancelReservationTx(reservation *Reservation) error {
	return r.getDB().Transaction(func(tx *gorm.DB) error {
		reservation.Status = ReservationStatusCancelled
		if err := tx.Save(reservation).Error; err != nil {
			return err
		}

		var roomData room.Room
		if err := tx.Where("id = ?", reservation.RoomID).First(&roomData).Error; err == nil {
			roomData.Status = room.RoomStatusAvailable
			if err := tx.Save(&roomData).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) UpdateReminderStatus(id uint, status string, sentAt *time.Time) error {
	updates := map[string]interface{}{
		"reminder_status":  status,
		"reminder_sent_at": sentAt,
	}
	return r.getDB().Model(&Reservation{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) FindUpcomingCheckoutsForReminder(windowEnd time.Time) ([]Reservation, error) {
	var reservations []Reservation
	err := r.getDB().
		Where("status = ? AND (reminder_sent_at IS NULL OR reminder_status = ?) AND check_out_date <= ?",
			ReservationStatusCheckIn, ReminderStatusPending, windowEnd).
		Preload("Room").
		Preload("User").
		Find(&reservations).Error
	if err != nil {
		return nil, err
	}
	return reservations, nil
}
