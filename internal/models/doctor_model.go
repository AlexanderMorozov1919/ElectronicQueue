package models

// DoctorStatus определяет статус врача
type DoctorStatus string

const (
	DoctorStatusActive   DoctorStatus = "активен"
	DoctorStatusInactive DoctorStatus = "неактивен"
	DoctorStatusOnBreak  DoctorStatus = "перерыв"
)

// Doctor представляет собой модель врача в базе данных.
type Doctor struct {
	ID             uint         `gorm:"primaryKey;autoIncrement;column:doctor_id" json:"id"`
	FullName       string       `gorm:"type:varchar(100);not null;column:full_name" json:"full_name"`
	Specialization string       `gorm:"type:varchar(100);not null" json:"specialization"`
	Login          string       `gorm:"column:login;unique" json:"login,omitempty"`
	PasswordHash   string       `gorm:"column:password_hash" json:"-"`
	Status         DoctorStatus `gorm:"type:varchar(20);default:'активен';column:status" json:"status"`
	Schedules      []Schedule   `gorm:"foreignKey:DoctorID;constraint:OnDelete:SET NULL" json:"schedules,omitempty"`
}

// DoctorResponse определяет данные, возвращаемые API.
type DoctorResponse struct {
	ID             uint         `json:"id"`
	FullName       string       `json:"full_name"`
	Specialization string       `json:"specialization"`
	Status         DoctorStatus `json:"status"`
}

// CreateDoctorRequest определяет структуру для создания нового врача.
type CreateDoctorRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Specialization string `json:"specialization" binding:"required"`
}

// UpdateDoctorRequest определяет структуру для обновления существующего врача.
type UpdateDoctorRequest struct {
	FullName       string        `json:"full_name,omitempty"`
	Specialization string        `json:"specialization,omitempty"`
	Status         *DoctorStatus `json:"status,omitempty"`
}
