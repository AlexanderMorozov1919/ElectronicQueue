package models

import (
	"encoding/base64"
	"time"
)

// Ad представляет рекламное объявление в базе данных.
type Ad struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Picture     []byte    `gorm:"type:bytea;not null" json:"-"`
	DurationSec int       `gorm:"column:duration_sec;not null;default:5" json:"duration_sec"`
	IsEnabled   bool      `gorm:"column:is_enabled;not null;default:true" json:"is_enabled"`
	ReceptionOn bool      `gorm:"column:reception_on;not null;default:true" json:"reception_on"`
	ScheduleOn  bool      `gorm:"column:schedule_on;not null;default:true" json:"schedule_on"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// AdResponse - DTO для ответа API, с картинкой в base64.
type AdResponse struct {
	ID          uint      `json:"id"`
	Picture     string    `json:"picture"`
	DurationSec int       `json:"duration_sec"`
	IsEnabled   bool      `json:"is_enabled"`
	ReceptionOn bool      `json:"reception_on"`
	ScheduleOn  bool      `json:"schedule_on"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAdRequest - DTO для создания объявления.
type CreateAdRequest struct {
	Picture     string `json:"picture" binding:"required"`
	DurationSec int    `json:"duration_sec" binding:"required,gt=0"`
	IsEnabled   bool   `json:"is_enabled"`
	ReceptionOn bool   `json:"reception_on"`
	ScheduleOn  bool   `json:"schedule_on"`
}

// UpdateAdRequest - DTO для обновления объявления.
type UpdateAdRequest struct {
	Picture     string `json:"picture,omitempty"`
	DurationSec *int   `json:"duration_sec,omitempty" binding:"omitempty,gt=0"`
	IsEnabled   *bool  `json:"is_enabled,omitempty"`
	ReceptionOn *bool  `json:"reception_on,omitempty"`
	ScheduleOn  *bool  `json:"schedule_on,omitempty"`
}

// ToResponse конвертирует модель Ad в AdResponse.
func (a *Ad) ToResponse() AdResponse {
	return AdResponse{
		ID:          a.ID,
		Picture:     base64.StdEncoding.EncodeToString(a.Picture),
		DurationSec: a.DurationSec,
		IsEnabled:   a.IsEnabled,
		ReceptionOn: a.ReceptionOn,
		ScheduleOn:  a.ScheduleOn,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
