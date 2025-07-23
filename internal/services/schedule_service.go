package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"gorm.io/gorm"
	"errors"
)

// ScheduleService предоставляет методы для управления расписаниями.
type ScheduleService struct {
	scheduleRepo repository.ScheduleRepository
	doctorRepo   repository.DoctorRepository // Для проверки существования врача
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