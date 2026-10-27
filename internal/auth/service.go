package auth

import (
	"errors"

	"hotel-app/utils"
)

var (
	ErrEmailAlreadyExists = errors.New("Email sudah digunakan oleh akun lain")
	ErrInvalidCredentials = errors.New("Email atau password tidak sesuai")
	ErrUserNotFound       = errors.New("Data pengguna tidak ditemukan")
	ErrPasswordHashing    = errors.New("Gagal memproses keamanan kata sandi")
	ErrDatabase           = errors.New("Gagal menyimpan akun ke database")
	ErrTokenGeneration    = errors.New("Gagal menerbitkan token otentikasi")
)

// Service mendefinisikan antarmuka logika bisnis untuk fitur otentikasi.
type Service interface {
	Register(input RegisterInput) (*UserResponse, error)
	Login(input LoginInput) (string, *UserResponse, error)
	GetCurrentUser(userID uint) (*UserResponse, error)
}

type service struct {
	repo Repository
}

// NewService membuat instance baru dari Auth Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Register(input RegisterInput) (*UserResponse, error) {
	// Cek apakah email sudah terdaftar
	existing, _ := s.repo.FindByEmail(input.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	user := User{
		Name:  input.Name,
		Email: input.Email,
	}

	if err := user.HashPassword(input.Password); err != nil {
		return nil, ErrPasswordHashing
	}

	if err := s.repo.Create(&user); err != nil {
		return nil, ErrDatabase
	}

	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *service) Login(input LoginInput) (string, *UserResponse, error) {
	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if !user.CheckPassword(input.Password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", nil, ErrTokenGeneration
	}

	return token, &UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *service) GetCurrentUser(userID uint) (*UserResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
