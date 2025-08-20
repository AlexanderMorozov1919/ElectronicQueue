package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type registrarPriorityRepo struct {
	db *gorm.DB
}

func NewRegistrarPriorityRepository(db *gorm.DB) RegistrarPriorityRepository {
	return &registrarPriorityRepo{db: db}
}

func (r *registrarPriorityRepo) GetPriorities(registrarID uint) ([]models.Service, error) {
	var services []models.Service
	err := r.db.Table("services").
		Joins("JOIN registrar_category_priorities rcp ON services.id = rcp.service_id").
		Where("rcp.registrar_id = ?", registrarID).
		Order("services.id ASC").
		Find(&services).Error
	return services, err
}

func (r *registrarPriorityRepo) SetPriorities(registrarID uint, serviceIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("registrar_id = ?", registrarID).Delete(&models.RegistrarCategoryPriority{}).Error; err != nil {
			return err
		}

		if len(serviceIDs) > 0 {
			priorities := make([]models.RegistrarCategoryPriority, len(serviceIDs))
			for i, serviceID := range serviceIDs {
				priorities[i] = models.RegistrarCategoryPriority{
					RegistrarID: registrarID,
					ServiceID:   serviceID,
				}
			}
			if err := tx.Create(&priorities).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
