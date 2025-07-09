package models

import (
	"time"
)

type TicketStatus string

// TicketStatus определяет перечисление для статусов талонов.
// @Description Перечисление статусов талонов электронной очереди
// @Enum string
// @swagger:model
// @Schema
const (
	StatusWaiting    TicketStatus = "ожидает"
	StatusInvited    TicketStatus = "приглашен" // Пациент вызван к окну
	StatusInProgress TicketStatus = "на_приеме"
	StatusCompleted  TicketStatus = "завершен"
	StatusRegistered TicketStatus = "зарегистрирован"
)

// Ticket представляет собой модель талона электронной очереди.
// @Description Модель талона электронной очереди
// @Name Ticket
// @swagger:model
// @Schema
type Ticket struct {
	ID              uint         `gorm:"primaryKey;autoIncrement;column:ticket_id" json:"id"`
	TicketNumber    string       `gorm:"type:varchar(20);not null;unique;column:ticket_number" json:"ticket_number"`
	Status          TicketStatus `gorm:"type:varchar(20);not null" json:"status"`
	ServiceCategory string       `gorm:"type:varchar(50);not null;column:service_category" json:"service_category"`
	WindowNumber    *int         `gorm:"column:window_number" json:"window_number,omitempty"`
	CreatedAt       time.Time    `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	CalledAt        *time.Time   `gorm:"column:called_at" json:"called_at,omitempty"`
	StartedAt       *time.Time   `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt     *time.Time   `gorm:"column:completed_at" json:"completed_at,omitempty"`
}

// TicketResponse определяет данные, возвращаемые API.
// @Description Ответ API с данными талона
// @Name TicketResponse
// @swagger:model
// @Schema
type TicketResponse struct {
	ID              uint         `json:"id"`
	TicketNumber    string       `json:"ticket_number"`
	Status          TicketStatus `json:"status"`
	ServiceCategory string       `json:"service_category"`
	WindowNumber    *int         `json:"window_number,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	CalledAt        *time.Time   `json:"called_at,omitempty"`
	StartedAt       *time.Time   `json:"started_at,omitempty"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty"`
}

// ToResponse преобразует модель Ticket в объект ответа TicketResponse (DTO)
func (t *Ticket) ToResponse() TicketResponse {
	return TicketResponse{
		ID:              t.ID,
		TicketNumber:    t.TicketNumber,
		Status:          t.Status,
		ServiceCategory: t.ServiceCategory,
		WindowNumber:    t.WindowNumber,
		CreatedAt:       t.CreatedAt,
		CalledAt:        t.CalledAt,
		StartedAt:       t.StartedAt,
		CompletedAt:     t.CompletedAt,
	}
}
