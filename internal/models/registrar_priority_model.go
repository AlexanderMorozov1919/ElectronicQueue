package models

// RegistrarCategoryPriority представляет join-таблицу для приоритетов регистратора.
type RegistrarCategoryPriority struct {
	RegistrarID uint `gorm:"primaryKey;column:registrar_id"`
	ServiceID   uint `gorm:"primaryKey;column:service_id"`
}

// TableName явно задает имя таблицы для GORM.
func (RegistrarCategoryPriority) TableName() string {
	return "registrar_category_priorities"
}
