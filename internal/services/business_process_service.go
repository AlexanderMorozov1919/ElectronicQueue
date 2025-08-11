package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"sync"
)

// BusinessProcessService управляет состоянием бизнес-процессов.
// Кэширует их в памяти для быстрой проверки в middleware.
type BusinessProcessService struct {
	repo       repository.BusinessProcessRepository
	log        *logger.AsyncLogger
	states     map[string]bool
	statesLock sync.RWMutex
}

func NewBusinessProcessService(repo repository.BusinessProcessRepository) (*BusinessProcessService, error) {
	service := &BusinessProcessService{
		repo:   repo,
		log:    logger.Default().WithField("module", "BusinessProcess"),
		states: make(map[string]bool),
	}
	if err := service.LoadProcesses(); err != nil {
		return nil, fmt.Errorf("failed to load initial business processes state: %w", err)
	}
	return service, nil
}

// LoadProcesses загружает все состояния из БД в кэш.
func (s *BusinessProcessService) LoadProcesses() error {
	s.statesLock.Lock()
	defer s.statesLock.Unlock()

	processes, err := s.repo.GetAll()
	if err != nil {
		s.log.WithError(err).Error("Failed to fetch business processes from DB")
		return err
	}

	for _, p := range processes {
		s.states[p.ProcessName] = p.IsEnabled
	}
	s.log.WithField("count", len(s.states)).Info("Business processes state loaded into memory")
	return nil
}

// IsEnabled проверяет, включен ли процесс. Безопасно для конкурентного доступа.
func (s *BusinessProcessService) IsEnabled(processName string) bool {
	s.statesLock.RLock()
	defer s.statesLock.RUnlock()
	enabled, exists := s.states[processName]
	return exists && enabled
}

// GetAll возвращает список всех процессов из БД.
func (s *BusinessProcessService) GetAll() ([]models.BusinessProcess, error) {
	return s.repo.GetAll()
}

// UpdateStatus обновляет состояние процесса в БД и в кэше.
func (s *BusinessProcessService) UpdateStatus(processName string, isEnabled bool) (*models.BusinessProcess, error) {
	process, err := s.repo.FindByName(processName)
	if err != nil {
		return nil, fmt.Errorf("process '%s' not found", processName)
	}

	process.IsEnabled = isEnabled
	if err := s.repo.Update(process); err != nil {
		return nil, fmt.Errorf("failed to update process '%s' in DB: %w", processName, err)
	}

	// Обновляем состояние в памяти только после успешного сохранения в БД
	s.statesLock.Lock()
	s.states[processName] = isEnabled
	s.statesLock.Unlock()

	s.log.WithField(processName, isEnabled).Info("Business process status updated")
	return process, nil
}
