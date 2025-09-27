package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type cleanupRepo struct {
	db *gorm.DB
}

func NewCleanupRepository(db *gorm.DB) CleanupRepository {
	return &cleanupRepo{db: db}
}

func (r *cleanupRepo) TruncateTickets() error {
	if err := r.db.Exec("TRUNCATE TABLE tickets RESTART IDENTITY CASCADE").Error; err != nil {
		return err
	}
	return nil
}

// GetTotalTicketsCount возвращает общее количество записей в таблице tickets.
func (r *cleanupRepo) GetTotalTicketsCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Ticket{}).Count(&count).Error
	return count, err
}
