package models

import (
	"time"
)

type TicketStatus string

// TicketStatus определяет перечисление для статусов талонов.
const (
	StatusWaiting    TicketStatus = "ожидает"
	StatusInvited    TicketStatus = "приглашен"
	StatusInProgress TicketStatus = "на_приеме"
	StatusCompleted  TicketStatus = "завершен"
	StatusRegistered TicketStatus = "зарегистрирован"
)

// Ticket представляет собой модель талона электронной очереди.
type Ticket struct {
	ID           uint         `gorm:"primaryKey;autoIncrement;column:ticket_id" json:"id"`
	TicketNumber string       `gorm:"type:varchar(20);not null;unique;column:ticket_number" json:"ticket_number"`
	Status       TicketStatus `gorm:"type:varchar(20);not null" json:"status"`
	ServiceType  *string      `gorm:"column:service_type" json:"service_type,omitempty"`
	WindowNumber *int         `gorm:"column:window_number" json:"window_number,omitempty"`
	QRCode       []byte       `gorm:"column:qr_code" json:"qr_code,omitempty"`
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
	ServiceType  *string      `json:"service_type,omitempty"`
	WindowNumber *int         `json:"window_number,omitempty"`
	QRCode       []byte       `json:"qr_code,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	CalledAt     *time.Time   `json:"called_at,omitempty"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}

// RegistrarTicketResponse расширяет Ticket, добавляя время записи для нужд регистратуры.
type RegistrarTicketResponse struct {
	Ticket
	AppointmentTime *string `json:"appointment_time,omitempty"`
}

// DoctorQueueTicketResponse определяет структуру для одного элемента в очереди к врачу.
type DoctorQueueTicketResponse struct {
	StartTime       string       `gorm:"column:start_time" json:"start_time"`
	TicketNumber    string       `gorm:"column:ticket_number" json:"ticket_number"`
	PatientFullName string       `gorm:"column:full_name" json:"patient_full_name"`
	Status          TicketStatus `gorm:"column:status" json:"status"`
}

// DailyReportRow представляет одну строку в ежедневном отчете по талонам.
type DailyReportRow struct {
	TicketNumber string `json:"ticket_number"`
	// PatientFullName      *string      `json:"patient_full_name"` // УДАЛЕНО
	DoctorFullName       *string      `json:"doctor_full_name"`
	DoctorSpecialization *string      `json:"doctor_specialization"`
	CabinetNumber        *int         `json:"cabinet_number"`
	AppointmentTime      *string      `json:"appointment_time"`
	Status               TicketStatus `json:"status"`
	CalledAt             *time.Time   `json:"called_at"`
	CompletedAt          *time.Time   `json:"completed_at"`
	Duration             *string      `json:"duration"`
}

// ToResponse преобразует модель Ticket в объект ответа TicketResponse (DTO)
func (t *Ticket) ToResponse() TicketResponse {
	return TicketResponse{
		ID:           t.ID,
		TicketNumber: t.TicketNumber,
		Status:       t.Status,
		ServiceType:  t.ServiceType,
		WindowNumber: t.WindowNumber,
		QRCode:       t.QRCode,
		CreatedAt:    t.CreatedAt,
		CalledAt:     t.CalledAt,
		StartedAt:    t.StartedAt,
		CompletedAt:  t.CompletedAt,
	}
}
