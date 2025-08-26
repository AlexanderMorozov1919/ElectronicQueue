package models

import (
	"time"
)

// Appointment представляет собой модель записи на прием (связь между пациентом, расписанием и талоном).
type Appointment struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:appointment_id" json:"id"`
	ScheduleID uint      `gorm:"not null;column:schedule_id" json:"schedule_id"`
	PatientID  *uint     `gorm:"column:patient_id" json:"patient_id,omitempty"`
	TicketID   *uint     `gorm:"column:ticket_id" json:"ticket_id,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	Patient    Patient   `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Schedule   Schedule  `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	Ticket     Ticket    `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
}

// CreateAppointmentRequest определяет структуру для создания новой записи на прием.
type CreateAppointmentRequest struct {
	ScheduleID uint  `json:"schedule_id" binding:"required"`
	PatientID  *uint `json:"patient_id"`
	TicketID   *uint `json:"ticket_id"`
}

// AppointmentResponse определяет данные, возвращаемые API.
type AppointmentResponse struct {
	ID        uint             `json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	Patient   PatientResponse  `json:"patient"`
	Schedule  ScheduleResponse `json:"schedule"`
}

// UpdateAppointmentRequest определяет структуру для добавления результатов приема.
type UpdateAppointmentRequest struct {
}

// ScheduleWithAppointmentInfo объединяет информацию о слоте расписания и записи на прием.
// Используется для отображения журнала-планировщика.
type ScheduleWithAppointmentInfo struct {
	Schedule
	Appointment  *Appointment `json:"appointment,omitempty"`
	TicketNumber *string      `json:"ticket_number,omitempty"`
}
