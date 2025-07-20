package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type serviceRepo struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepo{db: db}
}

func (r *serviceRepo) GetAll() ([]models.Service, error) {
	var services []models.Service
	if err := r.db.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

func (r *serviceRepo) GetByID(id uint) (*models.Service, error) {
	var service models.Service
	if err := r.db.First(&service, id).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepo) GetByServiceID(serviceID string) (*models.Service, error) {
	var service models.Service
	if err := r.db.Where("service_id = ?", serviceID).First(&service).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepo) Create(service *models.Service) error {
	return r.db.Create(service).Error
}

func (r *serviceRepo) Update(service *models.Service) error {
	return r.db.Save(service).Error
}

func (r *serviceRepo) Delete(id uint) error {
	return r.db.Delete(&models.Service{}, id).Error
}
