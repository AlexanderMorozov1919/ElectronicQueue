package repository

import (
	"ElectronicQueue/internal/models"
	"fmt"

	"gorm.io/gorm"
)

type adRepo struct {
	db *gorm.DB
}

func NewAdRepository(db *gorm.DB) AdRepository {
	return &adRepo{db: db}
}

func (r *adRepo) Create(ad *models.Ad) error {
	return r.db.Create(ad).Error
}

func (r *adRepo) GetAll() ([]models.Ad, error) {
	var ads []models.Ad
	if err := r.db.Order("created_at ASC").Find(&ads).Error; err != nil {
		return nil, err
	}
	return ads, nil
}

func (r *adRepo) GetEnabledFor(screen string) ([]models.Ad, error) {
	var ads []models.Ad
	query := r.db.Where("is_enabled = ?", true)

	switch screen {
	case "reception":
		query = query.Where("reception_on = ?", true)
	case "schedule":
		query = query.Where("schedule_on = ?", true)
	default:
		return nil, fmt.Errorf("unknown screen type: %s", screen)
	}

	if err := query.Order("id ASC").Find(&ads).Error; err != nil {
		return nil, err
	}
	return ads, nil
}

func (r *adRepo) GetByID(id uint) (*models.Ad, error) {
	var ad models.Ad
	if err := r.db.First(&ad, id).Error; err != nil {
		return nil, err
	}
	return &ad, nil
}

func (r *adRepo) Update(ad *models.Ad) error {
	return r.db.Save(ad).Error
}

func (r *adRepo) Delete(id uint) error {
	return r.db.Delete(&models.Ad{}, id).Error
}
