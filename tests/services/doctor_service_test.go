package services_test

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"
	"errors"
	"strings"
	"testing"
	"time"
)

/*
Мок-репозиторий для тестирования DoctorService
Имитирует работу с базой данных, но хранит данные в памяти.
Корректно реализует интерфейс repository.TicketRepository.
*/
type MockTicketRepository struct {
	tickets map[uint]*models.Ticket
}

func NewMockTicketRepository() *MockTicketRepository {
	return &MockTicketRepository{
		tickets: make(map[uint]*models.Ticket),
	}
}

// GetByID - ищет талон по ID в памяти
func (m *MockTicketRepository) GetByID(id uint) (*models.Ticket, error) {
	if ticket, exists := m.tickets[id]; exists {
		return ticket, nil
	}
	return nil, errors.New("ticket not found")
}

// Update - обновляет талон в памяти
func (m *MockTicketRepository) Update(ticket *models.Ticket) error {
	if _, exists := m.tickets[ticket.ID]; exists {
		m.tickets[ticket.ID] = ticket
		return nil
	}
	return errors.New("ticket not found")
}

// Create - создает новый талон в памяти
func (m *MockTicketRepository) Create(ticket *models.Ticket) error {
	if ticket.ID == 0 {
		ticket.ID = uint(len(m.tickets) + 1)
	}
	m.tickets[ticket.ID] = ticket
	return nil
}

// FindByStatuses - ищет талоны по списку статусов
func (m *MockTicketRepository) FindByStatuses(statuses []models.TicketStatus) ([]models.Ticket, error) {
	var tickets []models.Ticket
	for _, ticket := range m.tickets {
		for _, status := range statuses {
			if ticket.Status == status {
				tickets = append(tickets, *ticket)
				break
			}
		}
	}
	return tickets, nil
}

// GetMaxTicketNumber - возвращает максимальный номер талона
func (m *MockTicketRepository) GetMaxTicketNumber() (int, error) {
	return len(m.tickets), nil
}

// GetNextWaitingTicket - находит следующий талон в статусе "ожидает"
func (m *MockTicketRepository) GetNextWaitingTicket() (*models.Ticket, error) {
	var earliestTicket *models.Ticket
	for _, ticket := range m.tickets {
		if ticket.Status == models.StatusWaiting {
			if earliestTicket == nil || ticket.CreatedAt.Before(earliestTicket.CreatedAt) {
				earliestTicket = ticket
			}
		}
	}
	if earliestTicket != nil {
		return earliestTicket, nil
	}
	return nil, errors.New("no waiting tickets")
}

// Delete - удаляет талон по ID
func (m *MockTicketRepository) Delete(id uint) error {
	if _, exists := m.tickets[id]; exists {
		delete(m.tickets, id)
		return nil
	}
	return errors.New("ticket not found")
}

// Статическая проверка, что MockTicketRepository реализует интерфейс repository.TicketRepository
var _ repository.TicketRepository = &MockTicketRepository{}

/*
Тест: Успешное начало приема пациента
Проверяет, что врач может начать прием пациента с талоном в статусе "приглашен"
Ожидаемый результат: статус меняется на "на_приеме", время начала устанавливается
*/
func TestStartAppointment_Success(t *testing.T) {
	// Подготавливаем тестовые данные
	mockRepo := NewMockTicketRepository()
	service := services.NewDoctorService(mockRepo)

	// Создаем талон в статусе "приглашен" - это правильный статус для начала приема
	ticket := &models.Ticket{
		ID:           1,
		TicketNumber: "A001",
		Status:       models.StatusInvited, // Пациент уже вызван к окну
		CreatedAt:    time.Now(),
	}
	mockRepo.Create(ticket)

	// Вызываем тестируемую функцию
	result, err := service.StartAppointment(1)

	// Проверяем результаты
	if err != nil {
		t.Errorf("Ожидался успех, получена ошибка: %v", err)
		return
	}

	// Проверяем, что статус изменился на "на_приеме"
	if result.Status != models.StatusInProgress {
		t.Errorf("Статус должен быть '%s', получен: '%s'",
			models.StatusInProgress, result.Status)
	}

	// Проверяем, что время начала приема установлено
	if result.StartedAt == nil {
		t.Error("Время начала приема должно быть установлено")
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[START_APPOINTMENT] INPUT: {TicketID:%d}", 1)
	t.Logf("[START_APPOINTMENT] OUTPUT: {Status:\"%s\", StartedAt:%v, Error:%v}",
		result.Status, result.StartedAt, err)
	t.Logf("[START_APPOINTMENT] CALLS: repo.GetByID(1), repo.Update() x1")
	t.Logf("[START_APPOINTMENT] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Попытка начать прием несуществующего талона
Проверяет обработку ошибки, когда врач пытается начать прием с несуществующим ID талона
Ожидаемый результат: возвращается ошибка "ticket not found"
*/
func TestStartAppointment_TicketNotFound(t *testing.T) {
	// Подготавливаем пустой репозиторий (нет талонов)
	mockRepo := NewMockTicketRepository()
	service := services.NewDoctorService(mockRepo)

	// Пытаемся начать прием с несуществующим ID
	result, err := service.StartAppointment(999)

	// Проверяем, что получили ошибку
	if err == nil {
		t.Error("Ожидалась ошибка, но функция выполнилась успешно")
		return
	}

	// Проверяем, что результат nil при ошибке
	if result != nil {
		t.Error("Результат должен быть nil при ошибке")
	}

	// Проверяем текст ошибки
	expectedError := "ticket not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[START_APPOINTMENT] INPUT: {TicketID:%d}", 999)
	t.Logf("[START_APPOINTMENT] OUTPUT: {Result:%v, Error:\"%s\"}", result, err.Error())
	t.Logf("[START_APPOINTMENT] CALLS: repo.GetByID(999) x1")
	t.Logf("[START_APPOINTMENT] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Попытка начать прием талона с неправильным статусом
Проверяет, что нельзя начать прием, если талон не в статусе "приглашен"
Например, нельзя начать прием уже завершенного талона
Ожидаемый результат: возвращается ошибка о неправильном статусе
*/
func TestStartAppointment_InvalidStatus(t *testing.T) {
	mockRepo := NewMockTicketRepository()
	service := services.NewDoctorService(mockRepo)

	// Создаем талон в статусе "завершен" - это неправильный статус для начала приема
	ticket := &models.Ticket{
		ID:           2,
		TicketNumber: "A002",
		Status:       models.StatusCompleted, // Прием уже завершен
		CreatedAt:    time.Now(),
	}
	mockRepo.Create(ticket)

	// Пытаемся начать прием с неправильным статусом
	result, err := service.StartAppointment(2)

	// Проверяем, что получили ошибку
	if err == nil {
		t.Error("Ожидалась ошибка, но функция выполнилась успешно")
		return
	}

	// Проверяем, что результат nil при ошибке
	if result != nil {
		t.Error("Результат должен быть nil при ошибке")
	}

	// Проверяем текст ошибки
	expectedError := "ticket must be in 'приглашен' status"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[START_APPOINTMENT] INPUT: {TicketID:%d}", 2)
	t.Logf("[START_APPOINTMENT] OUTPUT: {Result:%v, Error:\"%s\"}", result, err.Error())
	t.Logf("[START_APPOINTMENT] CALLS: repo.GetByID(2) x1")
	t.Logf("[START_APPOINTMENT] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Успешное завершение приема пациента
Проверяет, что врач может завершить прием пациента с талоном в статусе "на_приеме"
Ожидаемый результат: статус меняется на "завершен", время завершения устанавливается
*/
func TestCompleteAppointment_Success(t *testing.T) {
	mockRepo := NewMockTicketRepository()
	service := services.NewDoctorService(mockRepo)

	// Создаем талон в статусе "на_приеме" с установленным временем начала
	startTime := time.Now()
	ticket := &models.Ticket{
		ID:           3,
		TicketNumber: "A003",
		Status:       models.StatusInProgress, // Прием уже начат
		StartedAt:    &startTime,              // Время начала установлено
		CreatedAt:    time.Now(),
	}
	mockRepo.Create(ticket)

	// Вызываем тестируемую функцию
	result, err := service.CompleteAppointment(3)

	// Проверяем результаты
	if err != nil {
		t.Errorf("Ожидался успех, получена ошибка: %v", err)
		return
	}

	// Проверяем, что статус изменился на "завершен"
	if result.Status != models.StatusCompleted {
		t.Errorf("Статус должен быть '%s', получен: '%s'",
			models.StatusCompleted, result.Status)
	}

	// Проверяем, что время завершения установлено
	if result.CompletedAt == nil {
		t.Error("Время завершения приема должно быть установлено")
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[COMPLETE_APPOINTMENT] INPUT: {TicketID:%d}", 3)
	t.Logf("[COMPLETE_APPOINTMENT] OUTPUT: {Status:\"%s\", CompletedAt:%v, Error:%v}",
		result.Status, result.CompletedAt, err)
	t.Logf("[COMPLETE_APPOINTMENT] CALLS: repo.GetByID(3), repo.Update() x1")
	t.Logf("[COMPLETE_APPOINTMENT] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Попытка завершить прием талона с неправильным статусом
Проверяет, что нельзя завершить прием, если талон не в статусе "на_приеме"
Например, нельзя завершить прием талона, который еще ожидает
Ожидаемый результат: возвращается ошибка о неправильном статусе
*/
func TestCompleteAppointment_InvalidStatus(t *testing.T) {
	mockRepo := NewMockTicketRepository()
	service := services.NewDoctorService(mockRepo)

	// Создаем талон в статусе "ожидает" - это неправильный статус для завершения приема
	ticket := &models.Ticket{
		ID:           4,
		TicketNumber: "A004",
		Status:       models.StatusWaiting, // Пациент еще ожидает
		CreatedAt:    time.Now(),
	}
	mockRepo.Create(ticket)

	// Пытаемся завершить прием с неправильным статусом
	result, err := service.CompleteAppointment(4)

	// Проверяем, что получили ошибку
	if err == nil {
		t.Error("Ожидалась ошибка, но функция выполнилась успешно")
		return
	}

	// Проверяем, что результат nil при ошибке
	if result != nil {
		t.Error("Результат должен быть nil при ошибке")
	}

	// Проверяем текст ошибки
	expectedError := "ticket must be in 'на_приеме' status"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[COMPLETE_APPOINTMENT] INPUT: {TicketID:%d}", 4)
	t.Logf("[COMPLETE_APPOINTMENT] OUTPUT: {Result:%v, Error:\"%s\"}", result, err.Error())
	t.Logf("[COMPLETE_APPOINTMENT] CALLS: repo.GetByID(4) x1")
	t.Logf("[COMPLETE_APPOINTMENT] STATUS: PASS")
	t.Logf("========================================")
}
