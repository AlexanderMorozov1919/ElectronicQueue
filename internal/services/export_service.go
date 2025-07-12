package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"reflect"
	"strings"
)

// allowedTables определяет, какие таблицы и поля доступны для экспорта.
var allowedTables = map[string][]string{
	"tickets":      {"ticket_id", "ticket_number", "status", "service_type", "window_number", "created_at", "called_at", "started_at", "completed_at"},
	"doctors":      {"doctor_id", "full_name", "specialization", "is_active"},
	"patients":     {"patient_id", "passport_series", "passport_number", "full_name", "birth_date", "phone", "oms_number"},
	"schedules":    {"schedule_id", "doctor_id", "date", "start_time", "end_time", "is_available"},
	"appointments": {"appointment_id", "schedule_id", "patient_id", "created_at"},
	"services":     {"id", "service_id", "name", "letter"},
}

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
	allowedColumns, ok := allowedTables[tableName]
	if !ok {
		return nil, 0, fmt.Errorf("table '%s' is not allowed for export", tableName)
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
		limit = 20 // Значение по умолчанию
	}
	if limit > 100 {
		limit = 100 // Максимальный лимит
	}

	return s.repo.GetData(tableName, page, limit, request.Filters)
}

// validateFilters проверяет корректность всех условий фильтрации.
func (s *ExportService) validateFilters(filters models.Filters, allowedColumns []string) error {
	op := strings.ToUpper(filters.LogicalOperator)
	if op != "AND" && op != "OR" && op != "" {
		return fmt.Errorf("invalid logical operator: %s", filters.LogicalOperator)
	}

	allowedOps := map[string]bool{
		"=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true, "LIKE": true, "IN": true,
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
			val := reflect.ValueOf(cond.Value)
			if val.Kind() != reflect.Slice {
				return fmt.Errorf("value for 'IN' operator on field '%s' must be an array", cond.Field)
			}
		}
	}
	return nil
}
