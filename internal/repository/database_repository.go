package repository

import (
	"ElectronicQueue/internal/models"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// DatabaseRepository определяет методы для работы с данными таблиц.
type DatabaseRepository interface {
	GetTableColumns(tableName string) ([]string, error)
	GetData(tableName string, page, limit int, filters models.Filters) ([]map[string]interface{}, int64, error)
	InsertData(tableName string, data interface{}) (int64, error)
	UpdateData(tableName string, data map[string]interface{}, filters models.Filters) (int64, error)
	DeleteData(tableName string, filters models.Filters) (int64, error)
}

type databaseRepo struct {
	db *gorm.DB
}

func NewDatabaseRepository(db *gorm.DB) DatabaseRepository {
	return &databaseRepo{db: db}
}

func (r *databaseRepo) applyFilters(tx *gorm.DB, filters models.Filters) (*gorm.DB, error) {
	if len(filters.Conditions) == 0 {
		return tx, nil
	}

	var queryParts []string
	var queryArgs []interface{}

	for _, cond := range filters.Conditions {
		isNil := cond.Value == nil
		op := strings.ToUpper(cond.Operator)

		var queryPart string
		if op == "IN" {
			queryPart = fmt.Sprintf("%s IN (?)", cond.Field)
			queryArgs = append(queryArgs, cond.Value)
		} else if isNil && (op == "=" || op == "IS") {
			queryPart = fmt.Sprintf("%s IS NULL", cond.Field)
		} else if isNil && (op == "<>" || op == "!=" || op == "IS NOT") {
			queryPart = fmt.Sprintf("%s IS NOT NULL", cond.Field)
		} else {
			queryPart = fmt.Sprintf("%s %s ?", cond.Field, cond.Operator)
			queryArgs = append(queryArgs, cond.Value)
		}
		queryParts = append(queryParts, queryPart)
	}

	logicalOp := " AND "
	if strings.ToUpper(filters.LogicalOperator) == "OR" {
		logicalOp = " OR "
	}

	fullQuery := strings.Join(queryParts, logicalOp)
	return tx.Where(fullQuery, queryArgs...), nil
}

// GetTableColumns получает список столбцов для указанной таблицы из схемы БД.
func (r *databaseRepo) GetTableColumns(tableName string) ([]string, error) {
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
func (r *databaseRepo) GetData(tableName string, page, limit int, filters models.Filters) ([]map[string]interface{}, int64, error) {
	tx := r.db.Table(tableName)

	// Построение WHERE-условия
	tx, err := r.applyFilters(tx, filters)
	if err != nil {
		return nil, 0, err
	}

	// Получение общего количества записей для пагинации
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применение пагинации
	offset := (page - 1) * limit
	tx = tx.Offset(offset).Limit(limit)

	// Если таблица - tickets, применяем кастомную сортировку
	if tableName == "tickets" {
		orderClause := "CASE status WHEN 'ожидает' THEN 1 WHEN 'приглашен' THEN 2 WHEN 'зарегистрирован' THEN 3 WHEN 'на_приеме' THEN 4 WHEN 'завершен' THEN 5 ELSE 6 END, created_at ASC"
		tx = tx.Order(orderClause)
	}

	// Выполнение запроса
	var results []map[string]interface{}
	if err := tx.Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// InsertData вставляет одну или несколько записей в таблицу.
func (r *databaseRepo) InsertData(tableName string, data interface{}) (int64, error) {
	// Начинаем транзакцию
	tx := r.db.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// Откатываем транзакцию в случае паники
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Передаем панику дальше
		}
	}()

	var totalRowsAffected int64

	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice:
		// Если это слайс, итерируем и вставляем по одному
		for i := 0; i < v.Len(); i++ {
			result := tx.Table(tableName).Create(v.Index(i).Interface())
			if result.Error != nil {
				tx.Rollback() // Откатываем транзакцию при ошибке
				return 0, result.Error
			}
			totalRowsAffected += result.RowsAffected
		}
	case reflect.Map:
		// Если это один объект (map), вставляем его
		result := tx.Table(tableName).Create(data)
		if result.Error != nil {
			tx.Rollback()
			return 0, result.Error
		}
		totalRowsAffected = result.RowsAffected
	default:
		tx.Rollback()
		return 0, fmt.Errorf("unsupported data type for insert: %T", data)
	}

	// Если все прошло успешно, коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return totalRowsAffected, nil
}

// UpdateData обновляет записи в таблице по заданным условиям.
func (r *databaseRepo) UpdateData(tableName string, data map[string]interface{}, filters models.Filters) (int64, error) {
	tx := r.db.Table(tableName)

	tx, err := r.applyFilters(tx, filters)
	if err != nil {
		return 0, err
	}

	result := tx.Updates(data)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *databaseRepo) DeleteData(tableName string, filters models.Filters) (int64, error) {
	tx := r.db.Table(tableName)

	tx, err := r.applyFilters(tx, filters)
	if err != nil {
		return 0, err
	}

	// Используем пустой map для GORM, чтобы он построил правильный DELETE запрос
	result := tx.Delete(&map[string]interface{}{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
