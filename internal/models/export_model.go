package models

// ExportRequest определяет тело запроса для экспорта данных.
type ExportRequest struct {
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
	Filters Filters `json:"filters"`
}

// InsertRequest определяет тело запроса для вставки данных.
// Поле Data может содержать один объект (map[string]interface{}) или массив объектов.
type InsertRequest struct {
	Data interface{} `json:"data" binding:"required"`
}

// UpdateRequest определяет тело запроса для обновления данных.
type UpdateRequest struct {
	Data    map[string]interface{} `json:"data" binding:"required"`
	Filters Filters                `json:"filters" binding:"required"`
}

// DeleteRequest определяет тело запроса для удаления данных.
type DeleteRequest struct {
	Filters Filters `json:"filters" binding:"required"`
}

// Filters содержит логический оператор и список условий для фильтрации.
type Filters struct {
	LogicalOperator string            `json:"logical_operator"`
	Conditions      []FilterCondition `json:"conditions"`
}

// FilterCondition описывает одно условие фильтрации.
type FilterCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}
