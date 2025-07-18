package models

// Doctor представляет собой модель врача в базе данных.
type Doctor struct {
	ID             uint       `gorm:"primaryKey;autoIncrement;column:doctor_id" json:"id"`
	FullName       string     `gorm:"type:varchar(100);not null;column:full_name" json:"full_name"`
	Specialization string     `gorm:"type:varchar(100);not null" json:"specialization"`
	IsActive       bool       `gorm:"default:true;column:is_active" json:"is_active"`
	Schedules      []Schedule `gorm:"foreignKey:DoctorID;constraint:OnDelete:SET NULL" json:"schedules,omitempty"`
}

// DoctorResponse определяет данные, возвращаемые API.
type DoctorResponse struct {
	ID             uint   `json:"id"`
	FullName       string `json:"full_name"`
	Specialization string `json:"specialization"`
	IsActive       bool   `json:"is_active"`
}

// CreateDoctorRequest определяет структуру для создания нового врача.
type CreateDoctorRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Specialization string `json:"specialization" binding:"required"`
}

// UpdateDoctorRequest определяет структуру для обновления существующего врача.
type UpdateDoctorRequest struct {
	FullName       string `json:"full_name,omitempty"`
	Specialization string `json:"specialization,omitempty"`
	IsActive       *bool  `json:"is_active,omitempty"`
}
