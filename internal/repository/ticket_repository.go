package repository

import (
	"ElectronicQueue/internal/models"
	"time"

	"gorm.io/gorm"
)

type ticketRepo struct {
	db *gorm.DB
}

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

func (r *ticketRepo) FindByStatus(status models.TicketStatus) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("status = ?", status).Order("created_at asc").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *ticketRepo) GetNextWaitingTicket() (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Where("status = ?", models.StatusWaiting).Order("created_at asc").First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetMaxTicketNumberForPrefix возвращает максимальный номер талона для конкретной буквы (префикса).
func (r *ticketRepo) GetMaxTicketNumberForPrefix(prefix string) (int, error) {
	var maxNum int
	query := `SELECT COALESCE(MAX(CAST(SUBSTRING(ticket_number, 2) AS INTEGER)), 0) FROM tickets WHERE ticket_number LIKE ?`
	err := r.db.Raw(query, prefix+"%").Scan(&maxNum).Error
	if err != nil {
		return 0, err
	}
	return maxNum, nil
}

func (r *ticketRepo) Delete(id uint) error {
	return r.db.Delete(&models.Ticket{}, id).Error
}

// FindInProgressTicketForCabinet находит талон в статусе "на приеме" для конкретного кабинета на сегодня.
func (r *ticketRepo) FindInProgressTicketForCabinet(cabinetNumber int) (*models.Ticket, error) {
	var ticket models.Ticket
	today := time.Now().Format("2006-01-02")

	err := r.db.Joins("JOIN appointments ON appointments.ticket_id = tickets.ticket_id").
		Joins("JOIN schedules ON schedules.schedule_id = appointments.schedule_id").
		Where("tickets.status = ? AND schedules.cabinet = ? AND schedules.date = ?",
			models.StatusInProgress, cabinetNumber, today).
		First(&ticket).Error

	if err != nil {
		return nil, err
	}
	return &ticket, nil
}
