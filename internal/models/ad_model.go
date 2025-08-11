package models

import (
	"encoding/base64"
	"time"
)

// Ad представляет рекламное объявление в базе данных.
type Ad struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Picture     []byte    `gorm:"type:bytea" json:"-"`
	Video       []byte    `gorm:"type:bytea" json:"-"`
	DurationSec int       `gorm:"column:duration_sec;not null;default:5" json:"duration_sec"`
	RepeatCount int       `gorm:"column:repeat_count;not null;default:1" json:"repeat_count"`
	IsEnabled   bool      `gorm:"column:is_enabled;not null;default:true" json:"is_enabled"`
	ReceptionOn bool      `gorm:"column:reception_on;not null;default:true" json:"reception_on"`
	ScheduleOn  bool      `gorm:"column:schedule_on;not null;default:true" json:"schedule_on"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// AdResponse - DTO для ответа API, с картинкой или видео в base64.
type AdResponse struct {
	ID          uint      `json:"id"`
	Picture     string    `json:"picture,omitempty"`
	Video       string    `json:"video,omitempty"`
	MediaType   string    `json:"media_type"`
	DurationSec int       `json:"duration_sec"`
	RepeatCount int       `json:"repeat_count"`
	IsEnabled   bool      `json:"is_enabled"`
	ReceptionOn bool      `json:"reception_on"`
	ScheduleOn  bool      `json:"schedule_on"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAdRequest - DTO для создания объявления.
type CreateAdRequest struct {
	Picture     string `json:"picture,omitempty"`
	Video       string `json:"video,omitempty"`
	DurationSec *int   `json:"duration_sec,omitempty" binding:"omitempty,gt=0"`
	RepeatCount *int   `json:"repeat_count,omitempty" binding:"omitempty,gt=0"`
	IsEnabled   bool   `json:"is_enabled"`
	ReceptionOn bool   `json:"reception_on"`
	ScheduleOn  bool   `json:"schedule_on"`
}

// UpdateAdRequest - DTO для обновления объявления.
type UpdateAdRequest struct {
	Picture     string `json:"picture,omitempty"`
	Video       string `json:"video,omitempty"`
	DurationSec *int   `json:"duration_sec,omitempty" binding:"omitempty,gt=0"`
	RepeatCount *int   `json:"repeat_count,omitempty" binding:"omitempty,gt=0"`
	IsEnabled   *bool  `json:"is_enabled,omitempty"`
	ReceptionOn *bool  `json:"reception_on,omitempty"`
	ScheduleOn  *bool  `json:"schedule_on,omitempty"`
}

// ToResponse конвертирует модель Ad в AdResponse.
func (a *Ad) ToResponse() AdResponse {
	resp := AdResponse{
		ID:          a.ID,
		DurationSec: a.DurationSec,
		RepeatCount: a.RepeatCount,
		IsEnabled:   a.IsEnabled,
		ReceptionOn: a.ReceptionOn,
		ScheduleOn:  a.ScheduleOn,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}

	if len(a.Video) > 0 {
		resp.MediaType = "video"
		resp.Video = base64.StdEncoding.EncodeToString(a.Video)
	} else if len(a.Picture) > 0 {
		resp.MediaType = "image"
		resp.Picture = base64.StdEncoding.EncodeToString(a.Picture)
	} else {
		resp.MediaType = "none"
	}

	return resp
}
