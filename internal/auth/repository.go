package auth

import (
	"hotel-app/config"

	"gorm.io/gorm"
)

// Repository mendefinisikan antarmuka akses data untuk fitur otentikasi.
type Repository interface {
	FindByEmail(email string) (*User, error)
	FindByID(id uint) (*User, error)
	Create(user *User) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository membuat instance baru dari Auth Repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return config.DB
}

func (r *repository) FindByEmail(email string) (*User, error) {
	var user User
	if err := r.getDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(id uint) (*User, error) {
	var user User
	if err := r.getDB().First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Create(user *User) error {
	return r.getDB().Create(user).Error
}
