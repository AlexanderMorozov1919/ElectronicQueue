package repository

import (
	"ElectronicQueue/internal/models"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ExportRepository определяет методы для экспорта данных.
type ExportRepository interface {
	GetData(tableName string, page, limit int, filters models.Filters) ([]map[string]interface{}, int64, error)
}

type exportRepo struct {
	db *gorm.DB
}

// NewExportRepository создает новый экземпляр ExportRepository.
func NewExportRepository(db *gorm.DB) ExportRepository {
	return &exportRepo{db: db}
}

// GetData строит и выполняет динамический запрос к БД.
func (r *exportRepo) GetData(tableName string, page, limit int, filters models.Filters) ([]map[string]interface{}, int64, error) {
	tx := r.db.Table(tableName)

	// Построение WHERE-условия
	if len(filters.Conditions) > 0 {
		var queryParts []string
		var queryArgs []interface{}

		for _, cond := range filters.Conditions {
			var queryPart string
			// GORM автоматически экранирует имена полей.
			if strings.ToUpper(cond.Operator) == "IN" {
				queryPart = fmt.Sprintf("%s IN (?)", cond.Field)
			} else {
				queryPart = fmt.Sprintf("%s %s ?", cond.Field, cond.Operator)
			}
			queryParts = append(queryParts, queryPart)
			queryArgs = append(queryArgs, cond.Value)
		}

		logicalOp := " AND "
		if strings.ToUpper(filters.LogicalOperator) == "OR" {
			logicalOp = " OR "
		}

		fullQuery := strings.Join(queryParts, logicalOp)
		tx = tx.Where(fullQuery, queryArgs...)
	}

	// Получение общего количества записей для пагинации
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применение пагинации
	offset := (page - 1) * limit
	tx = tx.Offset(offset).Limit(limit)

	// Выполнение запроса
	var results []map[string]interface{}
	if err := tx.Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
