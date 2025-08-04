package models

import "time"

// ReceptionLog представляет запись о времени обслуживания в регистратуре.
type ReceptionLog struct {
	LogID        uint           `gorm:"primaryKey;column:log_id"`
	TicketID     uint           `gorm:"not null;column:ticket_id"`
	RegistrarID  *uint          `gorm:"column:registrar_id"`
	WindowNumber int            `gorm:"not null;column:window_number"`
	CalledAt     time.Time      `gorm:"not null;column:called_at"`
	CompletedAt  *time.Time     `gorm:"column:completed_at"`
	Duration     *time.Duration `gorm:"column:duration"`
}
