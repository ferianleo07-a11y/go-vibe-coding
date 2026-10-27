package room

import (
	"errors"
	"fmt"
)

var (
	ErrRoomNotFound      = errors.New("Data kamar tidak ditemukan atau bukan milik Anda")
	ErrInvalidRoomStatus = errors.New("Status tidak valid. Pilihan: Tersedia, Terisi, Sedang Dibersihkan, Dalam Perbaikan")
	ErrRoomNumberExists  = errors.New("Nomor kamar sudah terdaftar di hotel Anda")
	ErrRoomOccupied      = errors.New("Kamar sedang terisi oleh tamu, selesaikan proses check-out terlebih dahulu sebelum menghapus")
)

// Service mendefinisikan antarmuka logika bisnis untuk entitas Kamar.
type Service interface {
	GetRooms(userID uint, status string) ([]Room, error)
	GetRoomByID(id uint, userID uint) (*Room, error)
	CreateRoom(userID uint, input RoomInput) (*Room, error)
	UpdateRoom(id uint, userID uint, input RoomInput) (*Room, error)
	DeleteRoom(id uint, userID uint) error
}

type service struct {
	repo Repository
}

// NewService membuat instance baru dari Room Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetRooms(userID uint, status string) ([]Room, error) {
	return s.repo.FindAllByUserID(userID, status)
}

func (s *service) GetRoomByID(id uint, userID uint) (*Room, error) {
	room, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

func (s *service) CreateRoom(userID uint, input RoomInput) (*Room, error) {
	status := input.Status
	if status == "" {
		status = RoomStatusAvailable
	} else if !IsValidRoomStatus(status) {
		return nil, ErrInvalidRoomStatus
	}

	existing, _ := s.repo.FindByRoomNumberAndUserID(input.RoomNumber, userID)
	if existing != nil {
		return nil, fmt.Errorf("Nomor kamar %s sudah terdaftar di hotel Anda", input.RoomNumber)
	}

	newRoom := Room{
		UserID:        userID,
		RoomNumber:    input.RoomNumber,
		Type:          input.Type,
		PricePerNight: input.PricePerNight,
		Status:        status,
		Description:   input.Description,
	}

	if err := s.repo.Create(&newRoom); err != nil {
		return nil, err
	}

	return &newRoom, nil
}

func (s *service) UpdateRoom(id uint, userID uint, input RoomInput) (*Room, error) {
	room, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	if input.Status != "" && !IsValidRoomStatus(input.Status) {
		return nil, ErrInvalidRoomStatus
	}

	if input.RoomNumber != room.RoomNumber {
		existing, _ := s.repo.FindByRoomNumberAndUserIDExcludingID(input.RoomNumber, userID, room.ID)
		if existing != nil {
			return nil, errors.New("Nomor kamar sudah digunakan oleh kamar lain")
		}
	}

	room.RoomNumber = input.RoomNumber
	room.Type = input.Type
	room.PricePerNight = input.PricePerNight
	if input.Status != "" {
		room.Status = input.Status
	}
	room.Description = input.Description

	if err := s.repo.Update(room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *service) DeleteRoom(id uint, userID uint) error {
	room, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return ErrRoomNotFound
	}

	if room.Status == RoomStatusOccupied {
		return ErrRoomOccupied
	}

	return s.repo.Delete(room)
}
