package models

// GetDataRequest определяет тело запроса для получения данных.
type GetDataRequest struct {
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
	Filters Filters `json:"filters"`
}

// Поле Data может содержать один объект (map[string]interface{}) или массив объектов.
type InsertRequest struct {
	Data interface{} `json:"data" binding:"required"`
}

type UpdateRequest struct {
	Data    map[string]interface{} `json:"data" binding:"required"`
	Filters Filters                `json:"filters" binding:"required"`
}

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
