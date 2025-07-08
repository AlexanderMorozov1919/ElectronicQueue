package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/utils"
	"fmt"
	"time"
)

const maxTicketNumber = 1000

// Service описывает услугу
// ID — уникальный идентификатор, Name — русское название
// Letter — буква для талона
type Service struct {
	ID     string
	Name   string
	Letter string
}

// TicketService предоставляет методы для работы с талонами
type TicketService struct {
	repo     repository.TicketRepository
	services []Service
}

// GetAllServices возвращает все доступные услуги (id, name, letter)
func (s *TicketService) GetAllServices() []Service {
	return s.services
}

// GetByID возвращает тикет по строковому id
func (s *TicketService) GetByID(idStr string) (*models.Ticket, error) {
	var id uint
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("GetByID: invalid id: %v", err))
		return nil, fmt.Errorf("invalid id")
	}
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("GetByID: repo error: %v", err))
		return nil, err
	}
	return ticket, nil
}

// CreateTicket создает новый талон для выбранной услуги
func (s *TicketService) CreateTicket(serviceID string) (*models.Ticket, error) {
	if serviceID == "" {
		logger.Default().Error("CreateTicket: serviceID is required")
		return nil, fmt.Errorf("serviceID is required")
	}
	ticketNumber, err := s.generateTicketNumber(serviceID)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("CreateTicket: failed to generate ticket number: %v", err))
		return nil, err
	}
	ticket := &models.Ticket{
		TicketNumber: ticketNumber,
		Status:       models.StatusWaiting,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Create(ticket); err != nil {
		logger.Default().Error(fmt.Sprintf("CreateTicket: repo create error: %v", err))
		return nil, err
	}
	return ticket, nil
}

// UpdateTicket обновляет тикет
func (s *TicketService) UpdateTicket(ticket *models.Ticket) error {
	err := s.repo.Update(ticket)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("UpdateTicket: repo update error: %v", err))
	}
	return err
}

// DeleteTicket удаляет тикет по строковому id
func (s *TicketService) DeleteTicket(idStr string) error {
	var id uint
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("DeleteTicket: invalid id: %v", err))
		return fmt.Errorf("invalid id")
	}
	err = s.repo.Delete(id)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("DeleteTicket: repo delete error: %v", err))
	}
	return err
}

// CallNextTicket вызывает следующего пациента в очереди к указанному окну
func (s *TicketService) CallNextTicket(windowNumber int) (*models.Ticket, error) {
	// Следующий талон для вызова
	ticket, err := s.repo.GetNextWaitingTicket()
	if err != nil {
		// Если талонов нет (gorm.ErrRecordNotFound), возвращаем ошибку, которую обработает хендлер
		logger.Default().Info(fmt.Sprintf("CallNextTicket: no waiting tickets in queue: %v", err))
		return nil, fmt.Errorf("очередь пуста")
	}

	// Обновление данных талона
	now := time.Now()
	ticket.Status = models.StatusInvited
	ticket.WindowNumber = &windowNumber
	ticket.CalledAt = &now

	// Сохраняем изменения в БД при вызове Update (сработает триггер и отправит NOTIFY)
	if err := s.repo.Update(ticket); err != nil {
		logger.Default().Error(fmt.Sprintf("CallNextTicket: repo update error: %v", err))
		return nil, err
	}

	logger.Default().Info(fmt.Sprintf("Ticket %s called to window %d", ticket.TicketNumber, windowNumber))
	return ticket, nil
}

// NewTicketService создает новый экземпляр TicketService
func NewTicketService(repo repository.TicketRepository) *TicketService {
	serviceList := []Service{
		{ID: "make_appointment", Name: "Записаться к врачу"},
		{ID: "confirm_appointment", Name: "Прием по записи"},
		{ID: "lab_tests", Name: "Сдать анализы"},
		{ID: "documents", Name: "Другой вопрос"},
	}
	alphabet := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i := range serviceList {
		if i < len(alphabet) {
			serviceList[i].Letter = string(alphabet[i])
		} else {
			serviceList[i].Letter = "Z"
		}
	}
	return &TicketService{repo: repo, services: serviceList}
}

// generateTicketNumber генерирует уникальный номер талона для услуги
func (s *TicketService) generateTicketNumber(serviceID string) (string, error) {
	letter := s.getServiceLetter(serviceID)
	maxNum, err := s.repo.GetMaxTicketNumber()
	if err != nil {
		logger.Default().Error(fmt.Sprintf("generateTicketNumber: repo error: %v", err))
		return "", err
	}
	num := maxNum + 1
	if num > maxTicketNumber {
		num = 1
	}
	return fmt.Sprintf("%s%03d", letter, num), nil
}

// getServiceLetter возвращает букву для услуги по её идентификатору
func (s *TicketService) getServiceLetter(serviceID string) string {
	for _, svc := range s.services {
		if svc.ID == serviceID {
			return svc.Letter
		}
	}
	return "Z"
}

// MapServiceIDToName возвращает название услуги по её идентификатору
func (s *TicketService) MapServiceIDToName(serviceID string) string {
	for _, svc := range s.services {
		if svc.ID == serviceID {
			return svc.Name
		}
	}
	return "Неизвестно"
}

// Модификация существующего метода для использования нового генератора
func (s *TicketService) GenerateTicketImage(baseSize int, ticket *models.Ticket, serviceName string) ([]byte, error) {
	data := utils.TicketData{
		ServiceName:  serviceName,
		TicketNumber: ticket.TicketNumber,
		DateTime:     ticket.CreatedAt,
	}

	// Данные для QR-кода
	qrData := []byte(fmt.Sprintf("Талон: %s\nВремя: %s\nУслуга: %s",
		ticket.TicketNumber,
		ticket.CreatedAt.Format("02.01.2006 15:04:05"),
		serviceName))

	// Генерируем изображение талона с заданным размером
	img, err := utils.GenerateTicketImageWithSizes(baseSize, qrData, data)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("GenerateTicketImage: failed to generate image: %v", err))
		return nil, err
	}
	return img, nil
}
