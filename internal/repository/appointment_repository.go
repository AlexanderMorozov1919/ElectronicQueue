package repository

import (
	"ElectronicQueue/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type appointmentRepo struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepo{db: db}
}

// CreateAppointmentInTransaction создает запись и блокирует слот в рамках одной транзакции.
func (r *appointmentRepo) CreateAppointmentInTransaction(req *models.CreateAppointmentRequest) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Блокируем строку расписания для безопасного обновления
		var schedule models.Schedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&schedule, req.ScheduleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("указанный слот в расписании не найден")
			}
			return err
		}

		// 2. Проверяем, свободен ли слот
		if !schedule.IsAvailable {
			return errors.New("выбранное время уже занято")
		}

		// 3. Создаем запись на прием
		appointment = models.Appointment{
			ScheduleID: req.ScheduleID,
			PatientID:  req.PatientID,
			TicketID:   req.TicketID,
		}
		if err := tx.Create(&appointment).Error; err != nil {
			return err
		}

		// 4. Обновляем статус слота на "занят"
		schedule.IsAvailable = false
		if err := tx.Save(&schedule).Error; err != nil {
			return err
		}

		// Если все успешно, транзакция будет автоматически закоммичена
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Загружаем связанные данные (пациент, расписание и врач) для полного ответа
	if err := r.db.Preload("Patient").Preload("Schedule.Doctor").First(&appointment, appointment.ID).Error; err != nil {
		return nil, err
	}

	return &appointment, nil
}

// FindScheduleAndAppointmentsByDoctorAndDate находит расписание и связанные с ним записи.
func (r *appointmentRepo) FindScheduleAndAppointmentsByDoctorAndDate(doctorID uint, date time.Time) ([]models.ScheduleWithAppointmentInfo, error) {
	var schedules []models.Schedule
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Находим все слоты расписания для врача на указанный день
	if err := r.db.Where("doctor_id = ? AND date >= ? AND date < ?", doctorID, startOfDay, endOfDay).Order("start_time asc").Find(&schedules).Error; err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return []models.ScheduleWithAppointmentInfo{}, nil
	}

	var result []models.ScheduleWithAppointmentInfo
	for _, s := range schedules {
		info := models.ScheduleWithAppointmentInfo{Schedule: s}
		// Если слот занят (is_available = false), находим связанную с ним запись
		if !s.IsAvailable {
			var app models.Appointment
			err := r.db.Preload("Patient").Where("schedule_id = ?", s.ID).First(&app).Error
			if err == nil {
				info.Appointment = &app
			}
		}
		result = append(result, info)
	}

	return result, nil
}
