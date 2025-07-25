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

// TruncateTickets удаляет завершенные tickets и связанные appointments
func (r *cleanupRepo) TruncateTickets() error {
	// Начинаем транзакцию
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Удаляем appointments с ticket_id = NULL (осиротевшие записи)
	if err := tx.Exec("DELETE FROM appointments WHERE ticket_id IS NULL").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Удаляем только tickets с completed_at != null
	if err := tx.Exec("DELETE FROM tickets WHERE completed_at IS NOT NULL").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Подтверждаем транзакцию
	return tx.Commit().Error
}

// GetTicketsCount возвращает количество завершенных записей в таблице tickets
func (r *cleanupRepo) GetTicketsCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Ticket{}).Where("completed_at IS NOT NULL").Count(&count).Error
	return count, err
}

// GetOrphanedAppointmentsCount возвращает количество appointments с ticket_id = NULL
func (r *cleanupRepo) GetOrphanedAppointmentsCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Appointment{}).Where("ticket_id IS NULL").Count(&count).Error
	return count, err
}
