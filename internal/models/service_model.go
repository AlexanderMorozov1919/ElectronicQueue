package models

type Service struct {
	ID        uint   `gorm:"primaryKey"`
	ServiceID string `gorm:"unique;not null"` // идентификатор для логики
	Name      string `gorm:"not null"`        // отображаемое имя
	Letter    string `gorm:"not null"`        // буква для талона
}
