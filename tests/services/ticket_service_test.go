package services_test

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/services"
	"errors"
	"strings"
	"testing"
	"time"
)

func init() {
	// Инициализируем логгер для тестов
	logger.Init("")
}

/*
Мок-репозиторий для тестирования услуг
Имитирует работу с таблицей услуг в базе данных.
Реализует интерфейс repository.ServiceRepository.
*/
type MockServiceRepository struct {
	services map[string]*models.Service
	nextID   uint
}

func NewMockServiceRepository() *MockServiceRepository {
	return &MockServiceRepository{
		services: make(map[string]*models.Service),
		nextID:   1,
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
	for _, service := range m.services {
		if service.ID == id {
			return service, nil
		}
	}
	return nil, errors.New("service not found")
}

// Create - создает новую услугу в памяти
func (m *MockServiceRepository) Create(service *models.Service) error {
	if service.ID == 0 {
		service.ID = m.nextID
		m.nextID++
	}
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

// Статическая проверка реализации интерфейса
var _ repository.ServiceRepository = &MockServiceRepository{}

/*
Расширенный мок-репозиторий для тестирования TicketService
Реализует интерфейс repository.TicketRepository.
*/
type MockTicketRepositoryWithServices struct {
	tickets map[uint]*models.Ticket
	maxNum  int
	nextID  uint
}

func NewMockTicketRepositoryWithServices() *MockTicketRepositoryWithServices {
	return &MockTicketRepositoryWithServices{
		tickets: make(map[uint]*models.Ticket),
		maxNum:  0,
		nextID:  1,
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
	if ticket.ID == 0 {
		ticket.ID = m.nextID
		m.nextID++
	}
	m.tickets[ticket.ID] = ticket
	m.maxNum++
	return nil
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

// Статическая проверка реализации интерфейса
var _ repository.TicketRepository = &MockTicketRepositoryWithServices{}

/*
Тест: Успешное создание талона
Проверяет, что система может создать новый талон для выбранной услуги
Ожидаемый результат: талон создается с уникальным номером и статусом "ожидает"
*/
func TestCreateTicket_Success(t *testing.T) {
	// Подготавливаем моки для репозиториев
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Добавляем тестовую услугу в мок
	service := &models.Service{
		ServiceID: "1",
		Name:      "Консультация",
		Letter:    "A",
	}
	mockServiceRepo.Create(service)

	// Создаем экземпляр сервиса с моками
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

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

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[CREATE_TICKET] INPUT: {ServiceID:\"%s\"}", "1")
	t.Logf("[CREATE_TICKET] OUTPUT: {TicketID:%d, Number:\"%s\", Status:\"%s\", Error:%v}",
		result.ID, result.TicketNumber, result.Status, err)
	t.Logf("[CREATE_TICKET] CALLS: repo.GetByServiceID(1), repo.GetMaxTicketNumber(), repo.Create()")
	t.Logf("[CREATE_TICKET] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Создание талона с неверным ID услуги
Проверяет обработку ошибки, когда пользователь выбирает несуществующую услугу
Ожидаемый результат: возвращается ошибка "service not found"
*/
func TestCreateTicket_InvalidServiceID(t *testing.T) {
	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис с пустым репозиторием услуг
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

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
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[CREATE_TICKET] INPUT: {ServiceID:\"%s\"}", "999")
	t.Logf("[CREATE_TICKET] OUTPUT: {Result:%v, Error:\"%s\"}", result, err.Error())
	t.Logf("[CREATE_TICKET] CALLS: repo.GetByServiceID(999) x1")
	t.Logf("[CREATE_TICKET] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Создание талона с пустым ID услуги
Проверяет валидацию входных данных
Ожидаемый результат: возвращается ошибка "serviceID is required"
*/
func TestCreateTicket_EmptyServiceID(t *testing.T) {
	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

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
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[CREATE_TICKET] INPUT: {ServiceID:\"\"}")
	t.Logf("[CREATE_TICKET] OUTPUT: {Result:%v, Error:\"%s\"}", result, err.Error())
	t.Logf("[CREATE_TICKET] CALLS: validation only")
	t.Logf("[CREATE_TICKET] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Успешный вызов следующего пациента
Проверяет, что система может вызвать следующего пациента из очереди
Ожидаемый результат: статус талона меняется на "приглашен", устанавливается номер окна и время вызова
*/
func TestCallNextTicket_Success(t *testing.T) {
	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем талон в статусе "ожидает" - это пациент в очереди
	ticket := &models.Ticket{
		TicketNumber: "A001",
		Status:       models.StatusWaiting, // Пациент ожидает в очереди
		CreatedAt:    time.Now(),
	}
	mockTicketRepo.Create(ticket)

	// Создаем сервис
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

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

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[CALL_NEXT_TICKET] INPUT: {WindowNumber:%d}", 1)
	t.Logf("[CALL_NEXT_TICKET] OUTPUT: {TicketID:%d, Number:\"%s\", Status:\"%s\", Window:%d, Error:%v}",
		result.ID, result.TicketNumber, result.Status, *result.WindowNumber, err)
	t.Logf("[CALL_NEXT_TICKET] CALLS: repo.GetNextWaitingTicket(), repo.Update()")
	t.Logf("[CALL_NEXT_TICKET] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Вызов пациента при пустой очереди
Проверяет обработку ситуации, когда в очереди нет пациентов
Ожидаемый результат: возвращается ошибка "очередь пуста"
*/
func TestCallNextTicket_EmptyQueue(t *testing.T) {
	// Подготавливаем моки с пустой очередью
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

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
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Ошибка должна содержать '%s', получено: '%s'",
			expectedError, err.Error())
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[CALL_NEXT_TICKET] INPUT: {WindowNumber:%d}", 1)
	t.Logf("[CALL_NEXT_TICKET] OUTPUT: {Result:%v, Error:\"%s\"}", result, err.Error())
	t.Logf("[CALL_NEXT_TICKET] CALLS: repo.GetNextWaitingTicket() x1")
	t.Logf("[CALL_NEXT_TICKET] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Получение всех услуг
Проверяет, что система может вернуть список всех доступных услуг
Ожидаемый результат: возвращается массив всех услуг
*/
func TestGetAllServices_Success(t *testing.T) {
	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Добавляем тестовые услуги
	mockServiceRepo.Create(&models.Service{ServiceID: "1", Name: "Консультация", Letter: "A"})
	mockServiceRepo.Create(&models.Service{ServiceID: "2", Name: "Анализы", Letter: "B"})

	// Создаем сервис
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

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

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[GET_ALL_SERVICES] INPUT: {}")
	t.Logf("[GET_ALL_SERVICES] OUTPUT: {Count:%d, Services:[%s, %s], Error:%v}",
		len(result), result[0].Name, result[1].Name, err)
	t.Logf("[GET_ALL_SERVICES] CALLS: repo.GetAll() x1")
	t.Logf("[GET_ALL_SERVICES] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Сопоставление ID услуги с названием
Проверяет, что система может найти название услуги по её ID
Ожидаемый результат: возвращается правильное название услуги
*/
func TestMapServiceIDToName_Success(t *testing.T) {
	// Подготавливаем моки
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Добавляем тестовую услугу
	mockServiceRepo.Create(&models.Service{ServiceID: "1", Name: "Консультация", Letter: "A"})

	// Создаем сервис
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

	// Ищем название услуги по ID
	result := serviceInstance.MapServiceIDToName("1")

	// Проверяем результат
	if result != "Консультация" {
		t.Errorf("Ожидалось 'Консультация', получено: '%s'", result)
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[MAP_SERVICE_ID_TO_NAME] INPUT: {ServiceID:\"%s\"}", "1")
	t.Logf("[MAP_SERVICE_ID_TO_NAME] OUTPUT: {Name:\"%s\"}", result)
	t.Logf("[MAP_SERVICE_ID_TO_NAME] CALLS: repo.GetByServiceID(1) x1")
	t.Logf("[MAP_SERVICE_ID_TO_NAME] STATUS: PASS")
	t.Logf("========================================")
}

/*
Тест: Сопоставление с неизвестной услугой
Проверяет обработку ситуации, когда запрашивается несуществующая услуга
Ожидаемый результат: возвращается "Неизвестно"
*/
func TestMapServiceIDToName_UnknownService(t *testing.T) {
	// Подготавливаем моки с пустым репозиторием услуг
	mockTicketRepo := NewMockTicketRepositoryWithServices()
	mockServiceRepo := NewMockServiceRepository()

	// Создаем сервис
	serviceInstance := services.NewTicketService(mockTicketRepo, mockServiceRepo)

	// Ищем название несуществующей услуги
	result := serviceInstance.MapServiceIDToName("999")

	// Проверяем результат
	if result != "Неизвестно" {
		t.Errorf("Ожидалось 'Неизвестно', получено: '%s'", result)
	}

	// Информативный вывод результата
	t.Logf("========================================")
	t.Logf("[MAP_SERVICE_ID_TO_NAME] INPUT: {ServiceID:\"%s\"}", "999")
	t.Logf("[MAP_SERVICE_ID_TO_NAME] OUTPUT: {Name:\"%s\"}", result)
	t.Logf("[MAP_SERVICE_ID_TO_NAME] CALLS: repo.GetByServiceID(999) x1")
	t.Logf("[MAP_SERVICE_ID_TO_NAME] STATUS: PASS")
	t.Logf("========================================")
}
