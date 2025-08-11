package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type receptionLogRepo struct {
	db *gorm.DB
}

func NewReceptionLogRepository(db *gorm.DB) ReceptionLogRepository {
	return &receptionLogRepo{db: db}
}

func (r *receptionLogRepo) Create(log *models.ReceptionLog) error {
	return r.db.Create(log).Error
}

func (r *receptionLogRepo) Update(log *models.ReceptionLog) error {
	return r.db.Save(log).Error
}

func (r *receptionLogRepo) FindActiveLogByTicketID(ticketID uint) (*models.ReceptionLog, error) {
	var log models.ReceptionLog
	err := r.db.Where("ticket_id = ? AND completed_at IS NULL", ticketID).First(&log).Error
	return &log, err
}
