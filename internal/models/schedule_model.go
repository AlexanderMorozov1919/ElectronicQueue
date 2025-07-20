package models

import (
	"time"
)

// Schedule представляет собой модель слота в расписании врача.
type Schedule struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:schedule_id" json:"schedule_id"`
	DoctorID    uint      `gorm:"not null;uniqueIndex:idx_doctor_date_start;column:doctor_id" json:"doctor_id"`
	Date        time.Time `gorm:"type:date;not null;uniqueIndex:idx_doctor_date_start" json:"date"`
	StartTime   string    `gorm:"type:time;not null;uniqueIndex:idx_doctor_date_start;column:start_time" json:"start_time"`
	EndTime     string    `gorm:"type:time;not null;column:end_time" json:"end_time"`
	IsAvailable bool      `gorm:"default:true;column:is_available" json:"is_available"`
	Doctor      Doctor    `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
}

// ScheduleResponse определяет данные, возвращаемые API, возможно с информацией о враче.
type ScheduleResponse struct {
	ID          uint      `json:"id"`
	DoctorID    uint      `json:"doctor_id"`
	Date        time.Time `json:"date"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	IsAvailable bool      `json:"is_available"`
}

// CreateScheduleRequest определяет структуру для создания нового слота в расписании.
type CreateScheduleRequest struct {
	DoctorID    uint      `json:"doctor_id" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	IsAvailable *bool     `json:"is_available"`
}

// UpdateScheduleRequest определяет структуру для обновления статуса слота (например, блокировка).
type UpdateScheduleRequest struct {
	IsAvailable *bool `json:"is_available" binding:"required"`
}
