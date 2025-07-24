package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ScheduleService предоставляет методы для управления расписаниями.
type ScheduleService struct {
	scheduleRepo repository.ScheduleRepository
	doctorRepo   repository.DoctorRepository
}

// NewScheduleService создает новый экземпляр ScheduleService.
func NewScheduleService(scheduleRepo repository.ScheduleRepository, doctorRepo repository.DoctorRepository) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		doctorRepo:   doctorRepo,
	}
}

// CreateSchedule создает новый слот в расписании.
func (s *ScheduleService) CreateSchedule(req *models.CreateScheduleRequest) (*models.Schedule, error) {
	_, err := s.doctorRepo.GetByID(req.DoctorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("врач с ID %d не найден", req.DoctorID)
		}
		return nil, fmt.Errorf("ошибка проверки врача: %w", err)
	}

	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	schedule := &models.Schedule{
		DoctorID:    req.DoctorID,
		Date:        req.Date,
		StartTime:   req.StartTime.Format("15:04:05"),
		EndTime:     req.EndTime.Format("15:04:05"),
		IsAvailable: isAvailable,
		Cabinet:     req.Cabinet,
	}

	if err := s.scheduleRepo.Create(schedule); err != nil {
		return nil, fmt.Errorf("не удалось создать слот в расписании: %w", err)
	}

	return schedule, nil
}

// DeleteSchedule удаляет слот из расписания по ID.
func (s *ScheduleService) DeleteSchedule(id uint) error {
	_, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("слот расписания с ID %d не найден", id)
		}
		return fmt.Errorf("ошибка при поиске слота расписания: %w", err)
	}

	if err := s.scheduleRepo.Delete(id); err != nil {
		return fmt.Errorf("не удалось удалить слот из расписания: %w", err)
	}
	return nil
}

// TodayScheduleResponse определяет структуру для ежедневного расписания.
type TodayScheduleResponse struct {
	Date         string                `json:"date"`
	MinStartTime string                `json:"min_start_time"`
	MaxEndTime   string                `json:"max_end_time"`
	Doctors      []DoctorScheduleModel `json:"doctors"`
}

// DoctorScheduleModel представляет расписание для одного врача.
type DoctorScheduleModel struct {
	ID             uint            `json:"id"`
	FullName       string          `json:"full_name"`
	Specialization string          `json:"specialization"`
	Slots          []TimeSlotModel `json:"slots"`
}

// TimeSlotModel представляет один временной слот в расписании.
type TimeSlotModel struct {
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	IsAvailable bool   `json:"is_available"`
	Cabinet     *int   `json:"cabinet,omitempty"`
}

// GetTodayScheduleState подготавливает данные для отображения дневного расписания.
func (s *ScheduleService) GetTodayScheduleState() (*TodayScheduleResponse, error) {
	today := time.Now()

	minTime, maxTime, err := s.scheduleRepo.FindMinMaxTimesForDate(today)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &TodayScheduleResponse{
				Date:         today.Format("2006-01-02"),
				MinStartTime: "09:00:00",
				MaxEndTime:   "18:00:00",
				Doctors:      []DoctorScheduleModel{},
			}, nil
		}
		return nil, fmt.Errorf("ошибка получения диапазона времени: %w", err)
	}

	allSchedules, err := s.scheduleRepo.FindAllSchedulesForDate(today)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения расписаний на сегодня: %w", err)
	}

	schedulesByDoctor := make(map[uint]DoctorScheduleModel)
	for _, schedule := range allSchedules {
		if schedule.Doctor.ID == 0 {
			continue
		}

		docSchedule, exists := schedulesByDoctor[schedule.DoctorID]
		if !exists {
			docSchedule = DoctorScheduleModel{
				ID:             schedule.Doctor.ID,
				FullName:       schedule.Doctor.FullName,
				Specialization: schedule.Doctor.Specialization,
				Slots:          []TimeSlotModel{},
			}
		}

		docSchedule.Slots = append(docSchedule.Slots, TimeSlotModel{
			StartTime:   schedule.StartTime,
			EndTime:     schedule.EndTime,
			IsAvailable: schedule.IsAvailable,
			Cabinet:     schedule.Cabinet,
		})
		schedulesByDoctor[schedule.DoctorID] = docSchedule
	}

	doctorSchedules := make([]DoctorScheduleModel, 0, len(schedulesByDoctor))
	for _, docSchedule := range schedulesByDoctor {
		doctorSchedules = append(doctorSchedules, docSchedule)
	}

	response := &TodayScheduleResponse{
		Date:         today.Format("2006-01-02"),
		MinStartTime: minTime.Format("15:04:05"),
		MaxEndTime:   maxTime.Format("15:04:05"),
		Doctors:      doctorSchedules,
	}

	return response, nil
}
