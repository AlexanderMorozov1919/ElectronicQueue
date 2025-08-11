package models

// Administrator представляет собой модель администратора в базе данных.
type Administrator struct {
	AdministratorID uint   `gorm:"primaryKey;column:administrator_id"`
	FullName        string `gorm:"type:varchar(100);not null;column:full_name"`
	Login           string `gorm:"column:login;unique"`
	PasswordHash    string `gorm:"column:password_hash"`
}
