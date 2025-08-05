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
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// AdResponse - DTO для ответа API, с картинкой в base64.
type AdResponse struct {
	ID          uint      `json:"id"`
	Picture     string    `json:"picture"` // base64 encoded
	DurationSec int       `json:"duration_sec"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAdRequest - DTO для создания объявления.
type CreateAdRequest struct {
	Picture     string `json:"picture" binding:"required"` // base64 encoded
	DurationSec int    `json:"duration_sec" binding:"required,gt=0"`
	IsEnabled   bool   `json:"is_enabled"`
}

// UpdateAdRequest - DTO для обновления объявления.
type UpdateAdRequest struct {
	Picture     string `json:"picture,omitempty"` // base64 encoded
	DurationSec *int   `json:"duration_sec,omitempty" binding:"omitempty,gt=0"`
	IsEnabled   *bool  `json:"is_enabled,omitempty"`
}

// ToResponse конвертирует модель Ad в AdResponse.
func (a *Ad) ToResponse() AdResponse {
	return AdResponse{
		ID:          a.ID,
		Picture:     base64.StdEncoding.EncodeToString(a.Picture),
		DurationSec: a.DurationSec,
		IsEnabled:   a.IsEnabled,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
