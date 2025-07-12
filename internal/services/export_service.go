package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"reflect"
	"strings"
)

// ExportService предоставляет методы для экспорта данных.
type ExportService struct {
	repo repository.ExportRepository
}

// NewExportService создает новый экземпляр ExportService.
func NewExportService(repo repository.ExportRepository) *ExportService {
	return &ExportService{repo: repo}
}

// ExportData выполняет валидацию и вызывает репозиторий для получения данных.
func (s *ExportService) ExportData(tableName string, request models.ExportRequest) ([]map[string]interface{}, int64, error) {
	allowedColumns, err := s.repo.GetTableColumns(tableName)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось проверить таблицу '%s': %w", tableName, err)
	}

	if err := s.validateFilters(request.Filters, allowedColumns); err != nil {
		return nil, 0, err
	}

	page := request.Page
	if page <= 0 {
		page = 1
	}
	limit := request.Limit
	if limit <= 0 {
		limit = 1000
	}

	return s.repo.GetData(tableName, page, limit, request.Filters)
}

// InsertData валидирует и вставляет данные.
func (s *ExportService) InsertData(tableName string, request models.InsertRequest) (int64, error) {
	_, err := s.repo.GetTableColumns(tableName)
	if err != nil {
		return 0, fmt.Errorf("не удалось проверить таблицу '%s': %w", tableName, err)
	}

	// TODO: Дополнительная валидация данных перед вставкой (например, проверка полей)
	// Для универсальности пока опускаем, но в продакшене это важно.

	return s.repo.InsertData(tableName, request.Data)
}

// UpdateData валидирует и обновляет данные.
func (s *ExportService) UpdateData(tableName string, request models.UpdateRequest) (int64, error) {
	allowedColumns, err := s.repo.GetTableColumns(tableName)
	if err != nil {
		return 0, fmt.Errorf("не удалось проверить таблицу '%s': %w", tableName, err)
	}

	// Критическая проверка: запрещаем обновление без фильтров
	if len(request.Filters.Conditions) == 0 {
		return 0, fmt.Errorf("обновление без фильтров запрещено")
	}

	if err := s.validateFilters(request.Filters, allowedColumns); err != nil {
		return 0, err
	}

	// Проверяем, что все поля в `data` существуют в таблице
	colsMap := make(map[string]bool)
	for _, col := range allowedColumns {
		colsMap[col] = true
	}
	for key := range request.Data {
		if !colsMap[key] {
			return 0, fmt.Errorf("поле '%s' не найдено в таблице '%s'", key, tableName)
		}
	}

	return s.repo.UpdateData(tableName, request.Data, request.Filters)
}

// DeleteData валидирует и удаляет данные.
func (s *ExportService) DeleteData(tableName string, request models.DeleteRequest) (int64, error) {
	allowedColumns, err := s.repo.GetTableColumns(tableName)
	if err != nil {
		return 0, fmt.Errorf("не удалось проверить таблицу '%s': %w", tableName, err)
	}

	// Критическая проверка: запрещаем удаление без фильтров
	if len(request.Filters.Conditions) == 0 {
		return 0, fmt.Errorf("удаление без фильтров запрещено")
	}

	if err := s.validateFilters(request.Filters, allowedColumns); err != nil {
		return 0, err
	}

	return s.repo.DeleteData(tableName, request.Filters)
}

// validateFilters проверяет корректность всех условий фильтрации.
func (s *ExportService) validateFilters(filters models.Filters, allowedColumns []string) error {
	op := strings.ToUpper(filters.LogicalOperator)
	if op != "AND" && op != "OR" && op != "" {
		return fmt.Errorf("invalid logical operator: %s", filters.LogicalOperator)
	}

	allowedOps := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true, "LIKE": true, "IN": true,
		"IS": true, "IS NOT": true,
	}

	colsMap := make(map[string]bool)
	for _, col := range allowedColumns {
		colsMap[col] = true
	}

	for _, cond := range filters.Conditions {
		if !colsMap[cond.Field] {
			return fmt.Errorf("field '%s' is not allowed for filtering in this table", cond.Field)
		}

		if !allowedOps[strings.ToUpper(cond.Operator)] {
			return fmt.Errorf("operator '%s' is not allowed", cond.Operator)
		}

		if strings.ToUpper(cond.Operator) == "IN" {
			if cond.Value == nil {
				return fmt.Errorf("value for 'IN' operator on field '%s' cannot be null", cond.Field)
			}
			val := reflect.ValueOf(cond.Value)
			if val.Kind() != reflect.Slice {
				return fmt.Errorf("value for 'IN' operator on field '%s' must be an array", cond.Field)
			}
		}
	}
	return nil
}
