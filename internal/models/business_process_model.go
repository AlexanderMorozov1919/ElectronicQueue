package models

// BusinessProcess представляет состояние бизнес-процесса в системе.
type BusinessProcess struct {
	ProcessName string `gorm:"primaryKey;column:process_name" json:"process_name"`
	IsEnabled   bool   `gorm:"column:is_enabled;not null"    json:"is_enabled"`
}
