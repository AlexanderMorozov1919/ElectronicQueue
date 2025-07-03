package services

import (
	"ElectronicQueue/internal/models/ticket_model"
	"ElectronicQueue/internal/repository"
	"time"

	"gorm.io/gorm"
)

type TicketsService struct {
	ticketRepo repository.TicketRepository
	db         *gorm.DB
}

func NewTicketsService(ticketRepo repository.TicketRepository, db *gorm.DB) *TicketsService {
	return &TicketsService{
		ticketRepo: ticketRepo,
		db:         db,
	}
}

// GetWaitingQueue возвращает все талоны со статусом 'ожидает', отсортированные по времени создания (FIFO)
func (s *TicketsService) GetWaitingQueue() ([]ticket_model.Ticket, error) {
	return s.ticketRepo.FindByStatuses([]ticket_model.TicketStatus{ticket_model.StatusWaiting})
}

// CallNextTicket вызывает следующего пациента (меняет статус и called_at)
func (s *TicketsService) CallNextTicket() (*ticket_model.Ticket, error) {
	var result *ticket_model.Ticket
	err := s.db.Transaction(func(tx *gorm.DB) error {
		tickets, err := s.ticketRepo.FindByStatuses([]ticket_model.TicketStatus{ticket_model.StatusWaiting})
		if err != nil {
			return err
		}
		if len(tickets) == 0 {
			return gorm.ErrRecordNotFound
		}
		ticket := tickets[0]
		now := time.Now()
		ticket.Status = ticket_model.StatusInvited
		ticket.CalledAt = &now
		if err := s.ticketRepo.Update(&ticket); err != nil {
			return err
		}
		result = &ticket
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
