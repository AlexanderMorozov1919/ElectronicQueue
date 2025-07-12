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
	GetTableColumns(tableName string) ([]string, error)
}

type exportRepo struct {
	db *gorm.DB
}

// NewExportRepository создает новый экземпляр ExportRepository.
func NewExportRepository(db *gorm.DB) ExportRepository {
	return &exportRepo{db: db}
}

// GetTableColumns получает список столбцов для указанной таблицы из схемы БД.
func (r *exportRepo) GetTableColumns(tableName string) ([]string, error) {
	var columns []string
	err := r.db.Raw(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = ?`,
		tableName,
	).Scan(&columns).Error

	if err != nil {
		return nil, fmt.Errorf("не удалось получить столбцы для таблицы %s: %w", tableName, err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("таблица '%s' не найдена или не имеет столбцов", tableName)
	}

	return columns, nil
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
