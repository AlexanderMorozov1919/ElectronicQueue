package models

// ExportRequest определяет тело запроса для экспорта данных.
type ExportRequest struct {
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
	Filters Filters `json:"filters"`
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
