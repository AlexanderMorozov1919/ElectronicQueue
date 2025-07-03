package repository

import (
	"ElectronicQueue/internal/models/ticket_model"

	"gorm.io/gorm"
)

type ticketRepo struct {
	db *gorm.DB
}

// NewTicketRepository - конструктор для ticketRepo.
func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepo{db: db}
}

func (r *ticketRepo) Create(ticket *ticket_model.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *ticketRepo) Update(ticket *ticket_model.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *ticketRepo) GetByID(id uint) (*ticket_model.Ticket, error) {
	var ticket ticket_model.Ticket
	if err := r.db.First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepo) FindByStatuses(statuses []ticket_model.TicketStatus) ([]ticket_model.Ticket, error) {
	var tickets []ticket_model.Ticket
	if err := r.db.Where("status IN ?", statuses).Order("created_at asc").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}
