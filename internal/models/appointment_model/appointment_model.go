package appointment_model

import (
	"time"

	"ElectronicQueue/internal/models/patient_model"
	"ElectronicQueue/internal/models/schedule_model"
	"ElectronicQueue/internal/models/ticket_model"
)

// Appointment представляет собой модель записи на прием (связь между пациентом, расписанием и талоном).
type Appointment struct {
	ID              uint                    `gorm:"primaryKey;autoIncrement;column:appointment_id" json:"id"`
	ScheduleID      uint                    `gorm:"not null;column:schedule_id" json:"schedule_id"`
	PatientID       uint                    `gorm:"not null;column:patient_id" json:"patient_id"`
	TicketID        uint                    `gorm:"not null;unique;column:ticket_id" json:"ticket_id"`
	Diagnosis       string                  `gorm:"type:text" json:"diagnosis,omitempty"`
	Recommendations string                  `gorm:"type:text" json:"recommendations,omitempty"`
	CreatedAt       time.Time               `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	Patient         patient_model.Patient   `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Schedule        schedule_model.Schedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	Ticket          ticket_model.Ticket     `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
}

// AppointmentResponse определяет данные, возвращаемые API.
type AppointmentResponse struct {
	ID              uint                            `json:"id"`
	Diagnosis       string                          `json:"diagnosis,omitempty"`
	Recommendations string                          `json:"recommendations,omitempty"`
	CreatedAt       time.Time                       `json:"created_at"`
	Patient         patient_model.PatientResponse   `json:"patient"`
	Schedule        schedule_model.ScheduleResponse `json:"schedule"`
	Ticket          ticket_model.TicketResponse     `json:"ticket"`
}

// CreateAppointmentRequest определяет структуру для создания новой записи на прием.
type CreateAppointmentRequest struct {
	ScheduleID uint `json:"schedule_id" binding:"required"`
	PatientID  uint `json:"patient_id" binding:"required"`
	TicketID   uint `json:"ticket_id" binding:"required"`
}

// UpdateAppointmentRequest определяет структуру для добавления результатов приема.
type UpdateAppointmentRequest struct {
	Diagnosis       string `json:"diagnosis,omitempty"`
	Recommendations string `json:"recommendations,omitempty"`
}
