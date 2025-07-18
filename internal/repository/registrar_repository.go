package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type registrarRepo struct {
	db *gorm.DB
}

func NewRegistrarRepository(db *gorm.DB) RegistrarRepository {
	return &registrarRepo{db: db}
}

func (r *registrarRepo) FindByLogin(login string) (*models.Registrar, error) {
	var registrar models.Registrar
	if err := r.db.Where("login = ?", login).First(&registrar).Error; err != nil {
		return nil, err
	}
	return &registrar, nil
}

func (r *registrarRepo) Create(registrar *models.Registrar) error {
	return r.db.Create(registrar).Error
}
