package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const maxTicketNumber = 1000

type TicketService struct {
	repo             repository.TicketRepository
	serviceRepo      repository.ServiceRepository
	receptionLogRepo repository.ReceptionLogRepository
	patientRepo      repository.PatientRepository
	appointmentRepo  repository.AppointmentRepository
	priorityRepo     repository.RegistrarPriorityRepository
	oneCService      *OneCService
}

func NewTicketService(
	repo repository.TicketRepository,
	serviceRepo repository.ServiceRepository,
	receptionLogRepo repository.ReceptionLogRepository,
	patientRepo repository.PatientRepository,
	appointmentRepo repository.AppointmentRepository,
	priorityRepo repository.RegistrarPriorityRepository,
	oneCService *OneCService,
) *TicketService {
	return &TicketService{
		repo:             repo,
		serviceRepo:      serviceRepo,
		receptionLogRepo: receptionLogRepo,
		patientRepo:      patientRepo,
		appointmentRepo:  appointmentRepo,
		priorityRepo:     priorityRepo,
		oneCService:      oneCService,
	}
}

func (s *TicketService) GetTicketsForRegistrar(categoryPrefix string, registrarID uint) ([]models.RegistrarTicketResponse, error) {
	statuses := []models.TicketStatus{
		models.StatusWaiting,
		models.StatusRegistered,
		models.StatusCompleted,
	}

	var prefixes []string
	if categoryPrefix != "" {
		prefixes = append(prefixes, categoryPrefix)
	} else {
		priorities, err := s.priorityRepo.GetPriorities(registrarID)
		if err != nil {
			return nil, err
		}
		if len(priorities) > 0 {
			for _, p := range priorities {
				prefixes = append(prefixes, p.Letter)
			}
		}
	}

	tickets, err := s.repo.FindForRegistrar(statuses, prefixes)
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
	isReceptionFinalStatus := ticket.Status == models.StatusCompleted || ticket.Status == models.StatusRegistered

	if isReceptionFinalStatus {
		return s.finalizeReceptionAndUpdateTicket(ticket)
	}

	err := s.repo.Update(ticket)
	if err != nil {
		logger.Default().WithError(err).Error(fmt.Sprintf("UpdateTicket: repo update error: %v", err))
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

func (s *TicketService) CallNextTicket(windowNumber int, categoryPrefix string, registrarID uint) (*models.Ticket, error) {
	var prefixes []string
	if categoryPrefix != "" {
		prefixes = append(prefixes, categoryPrefix)
	} else {
		priorities, err := s.priorityRepo.GetPriorities(registrarID)
		if err != nil {
			return nil, err
		}
		if len(priorities) > 0 {
			for _, p := range priorities {
				prefixes = append(prefixes, p.Letter)
			}
		}
	}

	ticket, err := s.repo.GetNextWaitingTicket(prefixes)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("очередь пуста")
		}
		return nil, err
	}

	now := time.Now()
	ticket.Status = models.StatusInvited
	ticket.WindowNumber = &windowNumber
	ticket.CalledAt = &now

	if err := s.repo.Update(ticket); err != nil {
		return nil, err
	}

	receptionLog := &models.ReceptionLog{
		TicketID:     ticket.ID,
		WindowNumber: windowNumber,
		CalledAt:     now,
		RegistrarID:  &registrarID,
	}
	if err := s.receptionLogRepo.Create(receptionLog); err != nil {
		logger.Default().WithError(err).Error("Failed to create reception log")
	}

	return ticket, nil
}

func (s *TicketService) CallSpecificTicket(ticketID uint, windowNumber int) (*models.Ticket, error) {
	ticket, err := s.repo.GetByID(ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("талон с ID %d не найден", ticketID)
		}
		return nil, fmt.Errorf("ошибка получения талона")
	}

	if ticket.Status != models.StatusWaiting {
		return nil, fmt.Errorf("талон %s имеет неверный статус '%s' для вызова (ожидался 'ожидает')", ticket.TicketNumber, ticket.Status)
	}

	now := time.Now()
	ticket.Status = models.StatusInvited
	ticket.WindowNumber = &windowNumber
	ticket.CalledAt = &now

	if err := s.repo.Update(ticket); err != nil {
		return nil, err
	}

	receptionLog := &models.ReceptionLog{
		TicketID:     ticket.ID,
		WindowNumber: windowNumber,
		CalledAt:     now,
	}
	if err := s.receptionLogRepo.Create(receptionLog); err != nil {
		logger.Default().WithError(err).Error("Failed to create reception log for specific call")
	}

	return ticket, nil
}

func (s *TicketService) GetInvitedTicketForWindow(windowNumber int) (*models.Ticket, error) {
	ticket, err := s.repo.FindInvitedByWindowNumber(windowNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Это не ошибка, просто нет активного талона
		}
		logger.Default().WithError(err).WithField("window", windowNumber).Error("GetInvitedTicketForWindow: repo error")
		return nil, err
	}
	return ticket, nil
}

func (s *TicketService) CheckInByPhone(phone string) (*models.Ticket, error) {
	data, err := s.oneCService.GetSchedule(phone)
	if err != nil {
		return nil, fmt.Errorf("ошибка сервиса 1С: %w", err)
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неверный формат ответа от сервиса 1С")
	}

	timeStr, ok := dataMap["appointment_time"].(string)
	if !ok {
		return nil, fmt.Errorf("поле 'appointment_time' не найдено или имеет неверный формат в ответе от сервиса 1С")
	}

	if timeStr == "" {
		return nil, errors.New("запись по данному номеру телефона не найдена на сегодня")
	}

	now := time.Now()
	appointmentTime, err := time.ParseInLocation("15:04:05", timeStr, now.Location())
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать время записи '%s': %w", timeStr, err)
	}

	fullAppointmentTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		appointmentTime.Hour(), appointmentTime.Minute(), appointmentTime.Second(), 0,
		now.Location(),
	)

	serviceID := "confirm_appointment"
	ticketNumber, err := s.generateTicketNumber(serviceID)
	if err != nil {
		return nil, err
	}

	newTicket := &models.Ticket{
		TicketNumber:    ticketNumber,
		Status:          models.StatusWaiting,
		CreatedAt:       time.Now(),
		ServiceType:     &serviceID,
		AppointmentTime: &fullAppointmentTime,
	}

	if err := s.repo.Create(newTicket); err != nil {
		return nil, fmt.Errorf("не удалось создать талон в базе данных: %w", err)
	}

	return newTicket, nil
}

func (s *TicketService) finalizeReceptionAndUpdateTicket(ticket *models.Ticket) error {
	log := logger.Default().WithField("ticket_id", ticket.ID)

	currentTicket, err := s.repo.GetByID(ticket.ID)
	if err != nil {
		log.WithError(err).Error("finalizeReception: failed to fetch current ticket state")
		return err
	}

	if currentTicket.Status == models.StatusRegistered && ticket.Status == models.StatusCompleted {
		log.Warn("finalizeReception: Attempted to mark a 'registered' ticket as 'completed'. Action ignored.")
		return nil
	}

	now := time.Now()

	if ticket.Status == models.StatusCompleted {
		ticket.CompletedAt = &now
	}

	if err := s.repo.Update(ticket); err != nil {
		log.WithError(err).Error("finalizeReception: failed to update ticket status")
		return err
	}

	receptionLog, err := s.receptionLogRepo.FindActiveLogByTicketID(ticket.ID)
	if err != nil {
		log.WithError(err).Warn("finalizeReception: active reception log not found, cannot stop timer")
		return nil
	}

	receptionLog.CompletedAt = &now
	duration := now.Sub(receptionLog.CalledAt)
	receptionLog.Duration = &duration

	if err := s.receptionLogRepo.Update(receptionLog); err != nil {
		log.WithError(err).Error("finalizeReception: failed to update reception log")
	}

	log.WithField("duration", duration).WithField("final_status", ticket.Status).Info("Reception finalized and logged")
	return nil
}

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
	if num >= maxTicketNumber {
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
