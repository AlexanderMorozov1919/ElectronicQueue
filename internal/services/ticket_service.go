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

// TicketService предоставляет методы для работы с талонами
type TicketService struct {
	repo        repository.TicketRepository
	serviceRepo repository.ServiceRepository
}

// NewTicketService создает новый экземпляр TicketService
func NewTicketService(repo repository.TicketRepository, serviceRepo repository.ServiceRepository) *TicketService {
	return &TicketService{repo: repo, serviceRepo: serviceRepo}
}

// GetTicketsForRegistrar получает список талонов для окна регистратора.
// Включает статусы: ожидает, зарегистрирован, завершен.
func (s *TicketService) GetTicketsForRegistrar(categoryPrefix string) ([]models.Ticket, error) {
	statuses := []models.TicketStatus{
		models.StatusWaiting,
		models.StatusRegistered,
		models.StatusCompleted,
	}
	tickets, err := s.repo.FindForRegistrar(statuses, categoryPrefix)
	if err != nil {
		logger.Default().WithError(err).Error("GetTicketsForRegistrar: repo error")
		return nil, err
	}
	return tickets, nil
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
	ticket := &models.Ticket{
		TicketNumber: ticketNumber,
		Status:       models.StatusWaiting,
		CreatedAt:    time.Now(),
		ServiceType:  &serviceID,
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

	now := time.Now()
	ticket.Status = models.StatusInvited
	ticket.WindowNumber = &windowNumber
	ticket.CalledAt = &now

	if err := s.repo.Update(ticket); err != nil {
		logger.Default().Error(fmt.Sprintf("CallNextTicket: repo update error: %v", err))
		return nil, err
	}

	logger.Default().Info(fmt.Sprintf("Ticket %s called to window %d", ticket.TicketNumber, windowNumber))
	return ticket, nil
}

// НОВЫЙ МЕТОД
func (s *TicketService) GetDailyReport() ([]models.DailyReportRow, error) {
	today := time.Now()
	report, err := s.repo.GetDailyReport(today)
	if err != nil {
		logger.Default().WithError(err).Error("GetDailyReport: service error")
		return nil, fmt.Errorf("ошибка получения данных для отчета: %w", err)
	}
	return report, nil
}

func (s *TicketService) generateTicketNumber(serviceID string) (string, error) {
	service, err := s.serviceRepo.GetByServiceID(serviceID)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("generateTicketNumber: service not found: %v", err))
		return "", err
	}
	letter := service.Letter
	maxNum, err := s.repo.GetMaxTicketNumberForPrefix(letter)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("generateTicketNumber: repo error getting max number for prefix %s: %v", letter, err))
		return "", err
	}

	num := maxNum + 1
	if num >= maxTicketNumber { // Use >= to be safe
		num = 1
	}
	return fmt.Sprintf("%s%03d", letter, num), nil
}

func (s *TicketService) MapServiceIDToName(serviceID string) string {
	service, err := s.serviceRepo.GetByServiceID(serviceID)
	if err != nil {
		return "Неизвестно"
	}
	return service.Name
}

func (s *TicketService) GenerateTicketImage(baseSize int, ticket *models.Ticket, serviceName string, mode string, qrData []byte) ([]byte, error) {
	waitingTickets, err := s.repo.FindByStatuses([]models.TicketStatus{models.StatusWaiting})
	waitingNumber := 0
	if err == nil {
		// Считаем только талоны, созданные до текущего
		for _, wt := range waitingTickets {
			if wt.CreatedAt.Before(ticket.CreatedAt) {
				waitingNumber++
			}
		}
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
