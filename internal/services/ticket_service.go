package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/utils"
	"fmt"
	"strings"
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
	repo        repository.TicketRepository
	serviceRepo repository.ServiceRepository
}

func (s *TicketService) GetAllServices() ([]models.Service, error) {
	return s.serviceRepo.GetAll()
}

func (s *TicketService) GetAllActiveTickets() ([]models.Ticket, error) {
	activeStatuses := []models.TicketStatus{models.StatusWaiting, models.StatusInvited}
	tickets, err := s.repo.FindByStatuses(activeStatuses)
	if err != nil {
		logger.Default().WithError(err).Error("GetAllActiveTickets: repo error")
		return nil, err
	}
	return tickets, nil
}

func (s *TicketService) GetServiceByID(id uint) (*models.Service, error) {
	return s.serviceRepo.GetByID(id)
}

func (s *TicketService) GetServiceByServiceID(serviceID string) (*models.Service, error) {
	return s.serviceRepo.GetByServiceID(serviceID)
}

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
	// Найти категорию услуги по serviceID
	// В качестве категории используем serviceID
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

func (s *TicketService) UpdateTicket(ticket *models.Ticket) error {
	err := s.repo.Update(ticket)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("UpdateTicket: repo update error: %v", err))
	}
	return err
}

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

func (s *TicketService) CallNextTicket(windowNumber int) (*models.Ticket, error) {
	ticket, err := s.repo.GetNextWaitingTicket()
	if err != nil {
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
func NewTicketService(repo repository.TicketRepository, serviceRepo repository.ServiceRepository) *TicketService {
	return &TicketService{repo: repo, serviceRepo: serviceRepo}
}

// generateTicketNumber генерирует уникальный номер талона для услуги
func (s *TicketService) generateTicketNumber(serviceID string) (string, error) {
	service, err := s.serviceRepo.GetByServiceID(serviceID)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("generateTicketNumber: service not found: %v", err))
		return "", err
	}
	letter := service.Letter
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

// MapServiceIDToName возвращает название услуги по её идентификатору
func (s *TicketService) MapServiceIDToName(serviceID string) string {
	service, err := s.serviceRepo.GetByServiceID(serviceID)
	if err != nil {
		return "Неизвестно"
	}
	return service.Name
}

// Модификация существующего метода для использования нового генератора
func (s *TicketService) GenerateTicketImage(baseSize int, ticket *models.Ticket, serviceName string, mode string, qrData []byte) ([]byte, error) {
	waitingTickets, err := s.repo.FindByStatuses([]models.TicketStatus{models.StatusWaiting})
	waitingNumber := len(waitingTickets) - 1
	if err != nil {
		waitingNumber = 0
	}

	background := "assets/img/ticket_bw.png"
	isColor := false
	if strings.ToLower(mode) == "color" {
		background = "assets/img/ticket.png"
		isColor = true
	}

	sqrt2 := 1.414
	width := int(float64(baseSize) / sqrt2)
	height := baseSize

	config := utils.TicketConfig{
		Width:          width,
		Height:         height,
		QRData:         qrData,
		FontPath:       "assets/fonts/Arial.ttf",
		BoldFontPath:   "assets/fonts/Arial_bold.ttf",
		BackgroundPath: background,
		ServiceName:    serviceName,
		TicketNumber:   ticket.TicketNumber,
		DateTime:       ticket.CreatedAt,
		WaitingNumber:  waitingNumber,
	}

	img, err := utils.GenerateTicketImage(config, isColor)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("GenerateTicketImage: failed to generate image: %v", err))
		return nil, err
	}
	return img, nil
}
