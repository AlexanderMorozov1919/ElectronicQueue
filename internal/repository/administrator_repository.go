package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type administratorRepo struct {
	db *gorm.DB
}

func NewAdministratorRepository(db *gorm.DB) AdministratorRepository {
	return &administratorRepo{db: db}
}

func (r *administratorRepo) FindByLogin(login string) (*models.Administrator, error) {
	var admin models.Administrator
	if err := r.db.Where("login = ?", login).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *administratorRepo) Create(admin *models.Administrator) error {
	return r.db.Create(admin).Error
}
