package models

import (
	"time"
)

// Patient представляет собой модель пациента в базе данных.
type Patient struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;column:patient_id" json:"id"`
	PassportSeries string    `gorm:"type:varchar(4);not null;uniqueIndex:idx_passport;column:passport_series" json:"passport_series"`
	PassportNumber string    `gorm:"type:varchar(6);not null;uniqueIndex:idx_passport;column:passport_number" json:"passport_number"`
	FullName       string    `gorm:"type:varchar(100);not null;column:full_name" json:"full_name"`
	BirthDate      time.Time `gorm:"type:date;column:birth_date" json:"birth_date"`
	Phone          string    `gorm:"type:varchar(20)" json:"phone"`
	OmsNumber      string    `gorm:"type:varchar(16);not null;column:oms_number" json:"oms_number"`
}

// PatientResponse определяет данные, возвращаемые API.
type PatientResponse struct {
	ID        uint      `json:"id"`
	FullName  string    `json:"full_name"`
	BirthDate time.Time `json:"birth_date"`
	Phone     string    `json:"phone"`
	OmsNumber string    `json:"oms_number"`
}

// CreatePatientRequest определяет структуру для создания нового пациента.
type CreatePatientRequest struct {
	PassportSeries string    `json:"passport_series" binding:"required,len=4"`
	PassportNumber string    `json:"passport_number" binding:"required,len=6"`
	FullName       string    `json:"full_name" binding:"required"`
	BirthDate      time.Time `json:"birth_date" binding:"required"`
	Phone          string    `json:"phone"`
	OmsNumber      string    `json:"oms_number" binding:"required,len=16"`
}

// UpdatePatientRequest определяет структуру для обновления существующего пациента.
type UpdatePatientRequest struct {
	PassportSeries string     `json:"passport_series,omitempty" binding:"omitempty,len=4"`
	PassportNumber string     `json:"passport_number,omitempty" binding:"omitempty,len=6"`
	FullName       string     `json:"full_name,omitempty"`
	BirthDate      *time.Time `json:"birth_date,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	OmsNumber      string     `json:"oms_number,omitempty" binding:"omitempty,len=16"`
}
