package models

import (
	"time"
)

// TicketStatus определяет перечисление для статусов талонов.
type TicketStatus string

const (
	StatusWaiting    TicketStatus = "ожидает"
	StatusInvited    TicketStatus = "приглашен"
	StatusInProgress TicketStatus = "на_приеме"
	StatusCompleted  TicketStatus = "завершен"
)

// Ticket представляет собой модель талона электронной очереди.
type Ticket struct {
	ID           uint         `gorm:"primaryKey;autoIncrement;column:ticket_id" json:"id"`
	TicketNumber string       `gorm:"type:varchar(20);not null;unique;column:ticket_number" json:"ticket_number"`
	Status       TicketStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt    time.Time    `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	CalledAt     *time.Time   `gorm:"column:called_at" json:"called_at,omitempty"`
	StartedAt    *time.Time   `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt  *time.Time   `gorm:"column:completed_at" json:"completed_at,omitempty"`
}

// TicketResponse определяет данные, возвращаемые API.
type TicketResponse struct {
	ID           uint         `json:"id"`
	TicketNumber string       `json:"ticket_number"`
	Status       TicketStatus `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	CalledAt     *time.Time   `json:"called_at,omitempty"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}
