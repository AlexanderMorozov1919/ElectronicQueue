package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type ticketRepo struct {
	db *gorm.DB
}

// NewTicketRepository - конструктор для ticketRepo.
func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepo{db: db}
}

func (r *ticketRepo) Create(ticket *models.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *ticketRepo) Update(ticket *models.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *ticketRepo) GetByID(id uint) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := r.db.First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepo) FindByStatuses(statuses []models.TicketStatus) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("status IN ?", statuses).Order("created_at asc").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

// GetMaxTicketNumber возвращает максимальный числовой номер билета (от 1 до 1000)
func (r *ticketRepo) GetMaxTicketNumber() (int, error) {
	var maxNum int
	// Предполагаем, что ticket_number всегда вида <буква><число>
	// Используем SUBSTRING для извлечения числа
	err := r.db.Raw(`SELECT COALESCE(MAX(CAST(SUBSTRING(ticket_number, 2) AS INTEGER)), 0) FROM tickets`).Scan(&maxNum).Error
	if err != nil {
		return 0, err
	}
	return maxNum, nil
}
