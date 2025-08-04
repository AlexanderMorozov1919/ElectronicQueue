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

// FindForRegistrar находит талоны по статусам и опционально по префиксу категории.
func (r *ticketRepo) FindForRegistrar(statuses []models.TicketStatus, categoryPrefix string) ([]models.Ticket, error) {
	var tickets []models.Ticket
	query := r.db.Where("status IN ?", statuses)

	if categoryPrefix != "" {
		query = query.Where("ticket_number LIKE ?", categoryPrefix+"%")
	}

	// Сортируем по убыванию времени создания, чтобы новые были вверху
	if err := query.Order("created_at DESC").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
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

func (r *ticketRepo) GetNextWaitingTicket(categoryPrefix string) (*models.Ticket, error) {
	var ticket models.Ticket
	query := r.db.Where("status = ?", models.StatusWaiting)

	if categoryPrefix != "" {
		query = query.Where("ticket_number LIKE ?", categoryPrefix+"%")
	}

	err := query.Order("created_at asc").First(&ticket).Error
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

// FindTicketsForCabinetQueue находит все талоны для очереди к кабинету врача.
// Возвращает список талонов со статусами 'на_приеме' и 'зарегистрирован'.
func (r *ticketRepo) FindTicketsForCabinetQueue(cabinetNumber int) ([]models.DoctorQueueTicketResponse, error) {
	var results []models.DoctorQueueTicketResponse
	today := time.Now().Format("2006-01-02")

	err := r.db.Table("tickets").
		Select("to_char(schedules.start_time, 'HH24:MI') as start_time, tickets.ticket_number, patients.full_name, tickets.status").
		Joins("JOIN appointments ON appointments.ticket_id = tickets.ticket_id").
		Joins("JOIN schedules ON schedules.schedule_id = appointments.schedule_id").
		Joins("JOIN patients ON patients.patient_id = appointments.patient_id").
		Where("schedules.cabinet = ? AND schedules.date = ? AND tickets.status IN ?",
			cabinetNumber, today, []string{string(models.StatusInProgress), string(models.StatusRegistered)}).
		Order("CASE WHEN tickets.status = 'на_приеме' THEN 0 ELSE 1 END, schedules.start_time ASC").
		Find(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}

// Найти талоны конкретного врача по статусу
// связь : талон - запись - расписание - врач
func (r *ticketRepo) FindByStatusAndDoctor(status models.TicketStatus, doctorID uint) ([]models.Ticket, error) {
	var tickets []models.Ticket
	err := r.db.Joins("JOIN appointments ON appointments.ticket_id = tickets.ticket_id").
		Joins("JOIN schedules ON schedules.schedule_id = appointments.schedule_id").
		Where("tickets.status = ? AND schedules.doctor_id = ?", status, doctorID).
		Order("schedules.start_time asc").
		Find(&tickets).Error
	return tickets, err
}

// GetDailyReport собирает данные для ежедневного отчета.
func (r *ticketRepo) GetDailyReport(date time.Time) ([]models.DailyReportRow, error) {
	var results []models.DailyReportRow

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Table("tickets as t").
		Select(`
            t.ticket_number,
            p.full_name as patient_full_name,
            d.full_name as doctor_full_name,
            d.specialization as doctor_specialization,
            s.cabinet as cabinet_number,
            to_char(s.start_time, 'HH24:MI') as appointment_time,
            t.status,
            rl.called_at,
            rl.completed_at,
            to_char(rl.duration, 'HH24:MI:SS') as duration
        `).
		Joins("LEFT JOIN appointments as a ON t.ticket_id = a.ticket_id").
		Joins("LEFT JOIN patients as p ON a.patient_id = p.patient_id").
		Joins("LEFT JOIN schedules as s ON a.schedule_id = s.schedule_id").
		Joins("LEFT JOIN doctors as d ON s.doctor_id = d.doctor_id").
		Joins("LEFT JOIN reception_logs as rl ON t.ticket_id = rl.ticket_id"). // ДОБАВЛЕНО
		Where("t.created_at >= ? AND t.created_at < ?", startOfDay, endOfDay).
		Order("t.created_at ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
