package models

import (
	"time"
)

// Appointment представляет собой модель записи на прием (связь между пациентом, расписанием и талоном).
type Appointment struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:appointment_id" json:"id"`
	ScheduleID uint      `gorm:"not null;column:schedule_id" json:"schedule_id"`
	PatientID  uint      `gorm:"not null;column:patient_id" json:"patient_id"`
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	Patient    Patient   `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Schedule   Schedule  `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	Ticket     Ticket    `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
}

// AppointmentResponse определяет данные, возвращаемые API.
type AppointmentResponse struct {
	ID        uint             `json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	Patient   PatientResponse  `json:"patient"`
	Schedule  ScheduleResponse `json:"schedule"`
}

// CreateAppointmentRequest определяет структуру для создания новой записи на прием.
type CreateAppointmentRequest struct {
	ScheduleID uint `json:"schedule_id" binding:"required"`
	PatientID  uint `json:"patient_id" binding:"required"`
}

// UpdateAppointmentRequest определяет структуру для добавления результатов приема.
type UpdateAppointmentRequest struct {
}
