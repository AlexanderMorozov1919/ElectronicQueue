package models

type Service struct {
	ID        uint   `gorm:"primaryKey" json:"-"`
	ServiceID string `gorm:"unique;not null" json:"id"`
	Name      string `gorm:"not null" json:"title"`
	Letter    string `gorm:"not null" json:"letter"`
}
