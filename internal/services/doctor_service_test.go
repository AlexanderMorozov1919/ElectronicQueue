package services

import (
	"ElectronicQueue/internal/models"
	"errors"
	"testing"
	"time"
)

/*
Мок-репозиторий для тестирования DoctorService
Имитирует работу с базой данных, но хранит данные в памяти
Это позволяет тестировать логику сервиса без реальной БД
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
	m.tickets[ticket.ID] = ticket
	return nil
}

// Create - создает новый талон в памяти
func (m *MockTicketRepository) Create(ticket *models.Ticket) error {
	m.tickets[ticket.ID] = ticket
	return nil
}

// GetAll - возвращает все талоны из памяти
func (m *MockTicketRepository) GetAll() ([]*models.Ticket, error) {
	tickets := make([]*models.Ticket, 0, len(m.tickets))
	for _, ticket := range m.tickets {
		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

// GetByStatus - ищет талоны по статусу
func (m *MockTicketRepository) GetByStatus(status string) ([]*models.Ticket, error) {
	var tickets []*models.Ticket
	for _, ticket := range m.tickets {
		if string(ticket.Status) == status {
			tickets = append(tickets, ticket)
		}
	}
	return tickets, nil
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
	for _, ticket := range m.tickets {
		if ticket.Status == models.StatusWaiting {
			return ticket, nil
		}
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

/*
Тест: Успешное начало приема пациента
Проверяет, что врач может начать прием пациента с талоном в статусе "приглашен"
Ожидаемый результат: статус меняется на "на_приеме", время начала устанавливается
*/
func TestStartAppointment_УспешноеНачалоПриема(t *testing.T) {
	t.Log("Тест: Успешное начало приема пациента")

	// Подготавливаем тестовые данные
	mockRepo := NewMockTicketRepository()
	service := NewDoctorService(mockRepo)

	// Создаем талон в статусе "приглашен" - это правильный статус для начала приема
	ticket := &models.Ticket{
		ID:           1,
		TicketNumber: "A001",
		Status:       models.StatusInvited, // Пациент уже вызван к окну
		CreatedAt:    time.Now(),
	}
	mockRepo.Create(ticket)

	t.Logf("Подготовлен талон: ID=%d, Номер=%s, Статус=%s",
		ticket.ID, ticket.TicketNumber, ticket.Status)

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

	t.Log("Тест успешно завершен")
}

/*
Тест: Попытка начать прием несуществующего талона
Проверяет обработку ошибки, когда врач пытается начать прием с несуществующим ID талона
Ожидаемый результат: возвращается ошибка "ticket not found"
*/
func TestStartAppointment_ТалонНеНайден(t *testing.T) {
	t.Log("Тест: Попытка начать прием несуществующего талона")

	// Подготавливаем пустой репозиторий (нет талонов)
	mockRepo := NewMockTicketRepository()
	service := NewDoctorService(mockRepo)

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
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	t.Log("Тест успешно завершен")
}

/*
Тест: Попытка начать прием талона с неправильным статусом
Проверяет, что нельзя начать прием, если талон не в статусе "приглашен"
Например, нельзя начать прием уже завершенного талона
Ожидаемый результат: возвращается ошибка о неправильном статусе
*/
func TestStartAppointment_НеверныйСтатус(t *testing.T) {
	t.Log("Тест: Попытка начать прием талона с неправильным статусом")

	mockRepo := NewMockTicketRepository()
	service := NewDoctorService(mockRepo)

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
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	t.Log("Тест успешно завершен")
}

/*
Тест: Успешное завершение приема пациента
Проверяет, что врач может завершить прием пациента с талоном в статусе "на_приеме"
Ожидаемый результат: статус меняется на "завершен", время завершения устанавливается
*/
func TestCompleteAppointment_УспешноеЗавершение(t *testing.T) {
	t.Log("Тест: Успешное завершение приема пациента")

	mockRepo := NewMockTicketRepository()
	service := NewDoctorService(mockRepo)

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

	t.Log("Тест успешно завершен")
}

/*
Тест: Попытка завершить прием талона с неправильным статусом
Проверяет, что нельзя завершить прием, если талон не в статусе "на_приеме"
Например, нельзя завершить прием талона, который еще ожидает
Ожидаемый результат: возвращается ошибка о неправильном статусе
*/
func TestCompleteAppointment_НеверныйСтатус(t *testing.T) {
	t.Log("Тест: Попытка завершить прием талона с неправильным статусом")

	mockRepo := NewMockTicketRepository()
	service := NewDoctorService(mockRepo)

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
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	t.Log("Тест успешно завершен")
}

// Вспомогательная функция для проверки содержимого строки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
