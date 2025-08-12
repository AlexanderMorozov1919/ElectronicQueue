package models

type Service struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ServiceID string `gorm:"unique;not null" json:"service_id"`
	Name      string `gorm:"not null" json:"title"`
	Letter    string `gorm:"not null" json:"letter"`
}
