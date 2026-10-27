package room

import (
	"hotel-app/config"

	"gorm.io/gorm"
)

// Repository mendefinisikan antarmuka akses data untuk entitas Kamar.
type Repository interface {
	FindAllByUserID(userID uint, status string) ([]Room, error)
	FindByIDAndUserID(id uint, userID uint) (*Room, error)
	FindByRoomNumberAndUserID(roomNumber string, userID uint) (*Room, error)
	FindByRoomNumberAndUserIDExcludingID(roomNumber string, userID uint, excludeID uint) (*Room, error)
	Create(room *Room) error
	Update(room *Room) error
	Delete(room *Room) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository membuat instance baru dari Room Repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return config.DB
}

func (r *repository) FindAllByUserID(userID uint, status string) ([]Room, error) {
	query := r.getDB().Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var rooms []Room
	if err := query.Order("room_number asc").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *repository) FindByIDAndUserID(id uint, userID uint) (*Room, error) {
	var room Room
	if err := r.getDB().Where("id = ? AND user_id = ?", id, userID).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *repository) FindByRoomNumberAndUserID(roomNumber string, userID uint) (*Room, error) {
	var room Room
	if err := r.getDB().Where("user_id = ? AND room_number = ?", userID, roomNumber).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *repository) FindByRoomNumberAndUserIDExcludingID(roomNumber string, userID uint, excludeID uint) (*Room, error) {
	var room Room
	if err := r.getDB().Where("user_id = ? AND room_number = ? AND id != ?", userID, roomNumber, excludeID).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *repository) Create(room *Room) error {
	return r.getDB().Create(room).Error
}

func (r *repository) Update(room *Room) error {
	return r.getDB().Save(room).Error
}

func (r *repository) Delete(room *Room) error {
	return r.getDB().Delete(room).Error
}
