package models

import "time"

type Registrar struct {
	RegistrarID  uint      `gorm:"primaryKey;column:registrar_id"`
	WindowNumber int       `gorm:"column:window_number;not null"`
	Login        string    `gorm:"column:login;unique;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	IsActive     bool      `gorm:"column:is_active;default:true"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}
