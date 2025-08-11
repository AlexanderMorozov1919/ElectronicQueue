package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type businessProcessRepo struct {
	db *gorm.DB
}

func NewBusinessProcessRepository(db *gorm.DB) BusinessProcessRepository {
	return &businessProcessRepo{db: db}
}

func (r *businessProcessRepo) GetAll() ([]models.BusinessProcess, error) {
	var processes []models.BusinessProcess
	if err := r.db.Find(&processes).Error; err != nil {
		return nil, err
	}
	return processes, nil
}

func (r *businessProcessRepo) Update(process *models.BusinessProcess) error {
	return r.db.Save(process).Error
}

func (r *businessProcessRepo) FindByName(name string) (*models.BusinessProcess, error) {
	var process models.BusinessProcess
	if err := r.db.Where("process_name = ?", name).First(&process).Error; err != nil {
		return nil, err
	}
	return &process, nil
}
