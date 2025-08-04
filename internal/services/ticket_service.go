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

// TicketService предоставляет методы для работы с талонами
type TicketService struct {
	repo             repository.TicketRepository
	serviceRepo      repository.ServiceRepository
	receptionLogRepo repository.ReceptionLogRepository
}

// NewTicketService создает новый экземпляр TicketService
func NewTicketService(
	repo repository.TicketRepository,
	serviceRepo repository.ServiceRepository,
	receptionLogRepo repository.ReceptionLogRepository,
) *TicketService {
	return &TicketService{
		repo:             repo,
		serviceRepo:      serviceRepo,
		receptionLogRepo: receptionLogRepo,
	}
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

// UpdateTicket обновляет талон и, если нужно, останавливает таймер обслуживания.
func (s *TicketService) UpdateTicket(ticket *models.Ticket) error {
	// НОВАЯ ЛОГИКА: Проверяем, является ли новый статус завершающим для этапа регистратуры
	isReceptionFinalStatus := ticket.Status == models.StatusCompleted || ticket.Status == models.StatusRegistered

	if isReceptionFinalStatus {
		// Если статус "завершен" или "зарегистрирован", вызываем специальную функцию
		return s.finalizeReceptionAndUpdateTicket(ticket)
	}

	// Для всех остальных статусов просто обновляем талон
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

func (s *TicketService) CallNextTicket(windowNumber int, categoryPrefix string) (*models.Ticket, error) {
	ticket, err := s.repo.GetNextWaitingTicket(categoryPrefix)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Default().WithField("category", categoryPrefix).Info("CallNextTicket: no waiting tickets in queue for category")
			return nil, fmt.Errorf("очередь пуста")
		}
		logger.Default().WithError(err).Error("CallNextTicket: repo error getting next ticket")
		return nil, err
	}

	now := time.Now()
	ticket.Status = models.StatusInvited
	ticket.WindowNumber = &windowNumber
	ticket.CalledAt = &now

	if err := s.repo.Update(ticket); err != nil {
		logger.Default().Error(fmt.Sprintf("CallNextTicket: repo update error: %v", err))
		return nil, err
	}

	// Создаем лог для таймера
	receptionLog := &models.ReceptionLog{
		TicketID:     ticket.ID,
		WindowNumber: windowNumber,
		CalledAt:     now,
	}
	if err := s.receptionLogRepo.Create(receptionLog); err != nil {
		// Логируем ошибку, но не прерываем основной процесс
		logger.Default().WithError(err).Error("Failed to create reception log")
	}

	logger.Default().Info(fmt.Sprintf("Ticket %s called to window %d", ticket.TicketNumber, windowNumber))
	return ticket, nil
}

// CallSpecificTicket вызывает конкретный талон по его ID.
func (s *TicketService) CallSpecificTicket(ticketID uint, windowNumber int) (*models.Ticket, error) {
	ticket, err := s.repo.GetByID(ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("талон с ID %d не найден", ticketID)
		}
		logger.Default().WithError(err).Error(fmt.Sprintf("CallSpecificTicket: repo error getting ticket by id %d", ticketID))
		return nil, fmt.Errorf("ошибка получения талона")
	}

	// Вызывать можно только талоны в статусе "ожидает"
	if ticket.Status != models.StatusWaiting {
		return nil, fmt.Errorf("талон %s имеет неверный статус '%s' для вызова (ожидался 'ожидает')", ticket.TicketNumber, ticket.Status)
	}

	now := time.Now()
	ticket.Status = models.StatusInvited
	ticket.WindowNumber = &windowNumber
	ticket.CalledAt = &now

	if err := s.repo.Update(ticket); err != nil {
		logger.Default().WithError(err).Error("CallSpecificTicket: repo update error")
		return nil, err
	}

	// Создаем лог для таймера
	receptionLog := &models.ReceptionLog{
		TicketID:     ticket.ID,
		WindowNumber: windowNumber,
		CalledAt:     now,
	}
	if err := s.receptionLogRepo.Create(receptionLog); err != nil {
		// Логируем ошибку, но не прерываем основной процесс
		logger.Default().WithError(err).Error("Failed to create reception log for specific call")
	}

	logger.Default().Info(fmt.Sprintf("Ticket %s specifically called to window %d", ticket.TicketNumber, windowNumber))
	return ticket, nil
}

// finalizeReceptionAndUpdateTicket обновляет статус талона и останавливает таймер в reception_logs.
func (s *TicketService) finalizeReceptionAndUpdateTicket(ticket *models.Ticket) error {
	log := logger.Default().WithField("ticket_id", ticket.ID)
	now := time.Now()

	// Устанавливаем время завершения в таблице tickets ТОЛЬКО для статуса "завершен".
	// Для "зарегистрирован" это поле остается пустым, так как талон уходит дальше.
	if ticket.Status == models.StatusCompleted {
		ticket.CompletedAt = &now
	}

	// 1. Обновляем сам талон в его таблице
	if err := s.repo.Update(ticket); err != nil {
		log.WithError(err).Error("finalizeReception: failed to update ticket status")
		return err
	}

	// 2. Ищем активный лог в reception_logs, чтобы его закрыть
	receptionLog, err := s.receptionLogRepo.FindActiveLogByTicketID(ticket.ID)
	if err != nil {
		log.WithError(err).Warn("finalizeReception: active reception log not found, cannot stop timer")
		return nil // Не возвращаем ошибку, т.к. основной процесс (смена статуса) прошел успешно
	}

	// 3. Обновляем лог: ставим время завершения и рассчитываем длительность
	receptionLog.CompletedAt = &now
	duration := now.Sub(receptionLog.CalledAt)
	receptionLog.Duration = &duration

	if err := s.receptionLogRepo.Update(receptionLog); err != nil {
		log.WithError(err).Error("finalizeReception: failed to update reception log")
		// Опять же, не возвращаем ошибку, чтобы не сломать клиент
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
