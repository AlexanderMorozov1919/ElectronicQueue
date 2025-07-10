package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"errors"
	"testing"
	"time"
)

func init() {
	// Инициализируем логгер для тестов
	logger.Init("")
}

/*
Мок-репозиторий для тестирования услуг
Имитирует работу с таблицей услуг в базе данных
Позволяет тестировать логику работы с услугами без реальной БД
*/
type MockServiceRepository struct {
	services map[string]*models.Service
}

func NewMockServiceRepository() *MockServiceRepository {
	return &MockServiceRepository{
		services: make(map[string]*models.Service),
	}
}

// GetAll - возвращает все услуги из памяти
func (m *MockServiceRepository) GetAll() ([]models.Service, error) {
	services := make([]models.Service, 0, len(m.services))
	for _, service := range m.services {
		services = append(services, *service)
	}
	return services, nil
}

// GetByID - ищет услугу по числовому ID
func (m *MockServiceRepository) GetByID(id uint) (*models.Service, error) {
	// Простая реализация для тестов
	for _, service := range m.services {
		if service.ID == id {
			return service, nil
		}
	}
	return nil, errors.New("service not found")
}

// Create - создает новую услугу в памяти
func (m *MockServiceRepository) Create(service *models.Service) error {
	m.services[service.ServiceID] = service
	return nil
}

// Update - обновляет услугу в памяти
func (m *MockServiceRepository) Update(service *models.Service) error {
	m.services[service.ServiceID] = service
	return nil
}

// Delete - удаляет услугу по числовому ID
func (m *MockServiceRepository) Delete(id uint) error {
	for serviceID, service := range m.services {
		if service.ID == id {
			delete(m.services, serviceID)
			return nil
		}
	}
	return errors.New("service not found")
}

// GetByServiceID - ищет услугу по строковому идентификатору
func (m *MockServiceRepository) GetByServiceID(serviceID string) (*models.Service, error) {
	if service, exists := m.services[serviceID]; exists {
		return service, nil
	}
	return nil, errors.New("service not found")
}

/*
Расширенный мок-репозиторий для тестирования TicketService
Включает дополнительные методы, необходимые для TicketService
*/
type MockTicketRepositoryWithServices struct {
	tickets map[uint]*models.Ticket
	maxNum  int
}

func NewMockTicketRepositoryWithServices() *MockTicketRepositoryWithServices {
	return &MockTicketRepositoryWithServices{
		tickets: make(map[uint]*models.Ticket),
		maxNum:  0,
	}
}

// GetByID - ищет талон по ID в памяти
func (m *MockTicketRepositoryWithServices) GetByID(id uint) (*models.Ticket, error) {
	if ticket, exists := m.tickets[id]; exists {
		return ticket, nil
	}
	return nil, errors.New("ticket not found")
}

// Update - обновляет талон в памяти
func (m *MockTicketRepositoryWithServices) Update(ticket *models.Ticket) error {
	m.tickets[ticket.ID] = ticket
	return nil
}

// Create - создает новый талон в памяти и увеличивает счетчик
func (m *MockTicketRepositoryWithServices) Create(ticket *models.Ticket) error {
	m.tickets[ticket.ID] = ticket
	m.maxNum++
	return nil
}

// GetAll - возвращает все талоны из памяти
func (m *MockTicketRepositoryWithServices) GetAll() ([]*models.Ticket, error) {
	tickets := make([]*models.Ticket, 0, len(m.tickets))
	for _, ticket := range m.tickets {
		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

// GetByStatus - ищет талоны по статусу
func (m *MockTicketRepositoryWithServices) GetByStatus(status string) ([]*models.Ticket, error) {
	var tickets []*models.Ticket
	for _, ticket := range m.tickets {
		if string(ticket.Status) == status {
			tickets = append(tickets, ticket)
		}
	}
	return tickets, nil
}

// Delete - удаляет талон по ID
func (m *MockTicketRepositoryWithServices) Delete(id uint) error {
	if _, exists := m.tickets[id]; exists {
		delete(m.tickets, id)
		return nil
	}
	return errors.New("ticket not found")
}

// GetMaxTicketNumber - возвращает максимальный номер талона
func (m *MockTicketRepositoryWithServices) GetMaxTicketNumber() (int, error) {
	return m.maxNum, nil
}

// GetNextWaitingTicket - находит следующий талон в статусе "ожидает"
func (m *MockTicketRepositoryWithServices) GetNextWaitingTicket() (*models.Ticket, error) {
	for _, ticket := range m.tickets {
		if ticket.Status == models.StatusWaiting {
			return ticket, nil
		}
	}
	return nil, errors.New("no waiting tickets")
}

// FindByStatuses - ищет талоны по списку статусов
func (m *MockTicketRepositoryWithServices) FindByStatuses(statuses []models.TicketStatus) ([]models.Ticket, error) {
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

/*
Тест: Успешное создание талона
Проверяет, что система может создать новый талон для выбранной услуги
Ожидаемый результат: талон создается с уникальным номером и статусом "ожидает"
*/
func TestCreateTicket_УспешноеСоздание(t *testing.T) {
	t.Log("Тест: Успешное создание талона")

	// Подготавливаем моки для репозиториев
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Добавляем тестовую услугу в мок
	service := &models.Service{
		ID:        1,
		ServiceID: "1",
		Name:      "Консультация",
		Letter:    "A",
	}
	mockServiceRepo.services["1"] = service

	// Создаем экземпляр сервиса с моками
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Вызываем тестируемую функцию
	result, err := serviceInstance.CreateTicket("1")

	// Проверяем результаты
	if err != nil {
		t.Errorf("Ожидался успех, получена ошибка: %v", err)
		return
	}

	// Проверяем, что талон создан
	if result == nil {
		t.Error("Результат не должен быть nil")
		return
	}

	// Проверяем, что статус установлен как "ожидает"
	if result.Status != models.StatusWaiting {
		t.Errorf("Статус должен быть '%s', получен: '%s'",
			models.StatusWaiting, result.Status)
	}

	// Проверяем, что номер талона сгенерирован
	if result.TicketNumber == "" {
		t.Error("Номер талона должен быть установлен")
	}

	t.Logf("Талон создан: Номер=%s, Статус=%s", result.TicketNumber, result.Status)
	t.Log("Тест успешно завершен")
}

/*
Тест: Создание талона с неверным ID услуги
Проверяет обработку ошибки, когда пользователь выбирает несуществующую услугу
Ожидаемый результат: возвращается ошибка "service not found"
*/
func TestCreateTicket_НеверныйServiceID(t *testing.T) {
	t.Log("Тест: Создание талона с неверным ID услуги")

	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис с пустым репозиторием услуг
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Пытаемся создать талон с несуществующим ID услуги
	result, err := serviceInstance.CreateTicket("999")

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
	expectedError := "service not found"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	t.Log("Тест успешно завершен")
}

/*
Тест: Создание талона с пустым ID услуги
Проверяет валидацию входных данных
Ожидаемый результат: возвращается ошибка "serviceID is required"
*/
func TestCreateTicket_ПустойServiceID(t *testing.T) {
	t.Log("Тест: Создание талона с пустым ID услуги")

	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Пытаемся создать талон с пустым ID услуги
	result, err := serviceInstance.CreateTicket("")

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
	expectedError := "serviceID is required"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	t.Log("Тест успешно завершен")
}

/*
Тест: Успешный вызов следующего пациента
Проверяет, что система может вызвать следующего пациента из очереди
Ожидаемый результат: статус талона меняется на "приглашен", устанавливается номер окна и время вызова
*/
func TestCallNextTicket_УспешныйВызов(t *testing.T) {
	t.Log("Тест: Успешный вызов следующего пациента")

	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем талон в статусе "ожидает" - это пациент в очереди
	ticket := &models.Ticket{
		ID:           1,
		TicketNumber: "A001",
		Status:       models.StatusWaiting, // Пациент ожидает в очереди
		CreatedAt:    time.Now(),
	}
	mockTicketRepo.Create(ticket)

	// Создаем сервис
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Вызываем следующего пациента к окну №1
	result, err := serviceInstance.CallNextTicket(1)

	// Проверяем результаты
	if err != nil {
		t.Errorf("Ожидался успех, получена ошибка: %v", err)
		return
	}

	// Проверяем, что статус изменился на "приглашен"
	if result.Status != models.StatusInvited {
		t.Errorf("Статус должен быть '%s', получен: '%s'",
			models.StatusInvited, result.Status)
	}

	// Проверяем, что номер окна установлен
	if result.WindowNumber == nil || *result.WindowNumber != 1 {
		t.Error("Номер окна должен быть установлен")
	}

	// Проверяем, что время вызова установлено
	if result.CalledAt == nil {
		t.Error("Время вызова должно быть установлено")
	}

	t.Logf("Пациент вызван: Номер=%s, Окно=%d", result.TicketNumber, *result.WindowNumber)
	t.Log("Тест успешно завершен")
}

/*
Тест: Вызов пациента при пустой очереди
Проверяет обработку ситуации, когда в очереди нет пациентов
Ожидаемый результат: возвращается ошибка "очередь пуста"
*/
func TestCallNextTicket_ОчередьПуста(t *testing.T) {
	t.Log("Тест: Вызов пациента при пустой очереди")

	// Подготавливаем моки с пустой очередью
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Пытаемся вызвать пациента из пустой очереди
	result, err := serviceInstance.CallNextTicket(1)

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
	expectedError := "очередь пуста"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	t.Log("Тест успешно завершен")
}

/*
Тест: Получение всех услуг
Проверяет, что система может вернуть список всех доступных услуг
Ожидаемый результат: возвращается массив всех услуг
*/
func TestGetAllServices_УспешноеПолучение(t *testing.T) {
	t.Log("Тест: Получение всех услуг")

	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Добавляем тестовые услуги
	service1 := &models.Service{ID: 1, ServiceID: "1", Name: "Консультация", Letter: "A"}
	service2 := &models.Service{ID: 2, ServiceID: "2", Name: "Анализы", Letter: "B"}
	mockServiceRepo.services["1"] = service1
	mockServiceRepo.services["2"] = service2

	// Создаем сервис
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Получаем список всех услуг
	result, err := serviceInstance.GetAllServices()

	// Проверяем результаты
	if err != nil {
		t.Errorf("Ожидался успех, получена ошибка: %v", err)
		return
	}

	// Проверяем количество услуг
	if len(result) != 2 {
		t.Errorf("Ожидалось 2 услуги, получено: %d", len(result))
	}

	t.Logf("Получено услуг: %d", len(result))
	t.Log("Тест успешно завершен")
}

/*
Тест: Сопоставление ID услуги с названием
Проверяет, что система может найти название услуги по её ID
Ожидаемый результат: возвращается правильное название услуги
*/
func TestMapServiceIDToName_УспешноеСопоставление(t *testing.T) {
	t.Log("Тест: Сопоставление ID услуги с названием")

	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Добавляем тестовую услугу
	service := &models.Service{ID: 1, ServiceID: "1", Name: "Консультация", Letter: "A"}
	mockServiceRepo.services["1"] = service

	// Создаем сервис
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Ищем название услуги по ID
	result := serviceInstance.MapServiceIDToName("1")

	// Проверяем результат
	if result != "Консультация" {
		t.Errorf("Ожидалось 'Консультация', получено: '%s'", result)
	}

	t.Logf("Название услуги: %s", result)
	t.Log("Тест успешно завершен")
}

/*
Тест: Сопоставление с неизвестной услугой
Проверяет обработку ситуации, когда запрашивается несуществующая услуга
Ожидаемый результат: возвращается "Неизвестно"
*/
func TestMapServiceIDToName_НеизвестнаяУслуга(t *testing.T) {
	t.Log("Тест: Сопоставление с неизвестной услугой")

	// Подготавливаем моки с пустым репозиторием услуг
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис
	serviceInstance := NewTicketService(mockTicketRepo, mockServiceRepo)

	// Ищем название несуществующей услуги
	result := serviceInstance.MapServiceIDToName("999")

	// Проверяем результат
	if result != "Неизвестно" {
		t.Errorf("Ожидалось 'Неизвестно', получено: '%s'", result)
	}

	t.Logf("Результат для неизвестной услуги: %s", result)
	t.Log("Тест успешно завершен")
}
