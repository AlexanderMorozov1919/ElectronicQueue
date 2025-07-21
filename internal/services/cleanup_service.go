package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/repository"

	"github.com/sirupsen/logrus"
)

type CleanupService struct {
	repo repository.CleanupRepository
	log  *logger.AsyncLogger
}

func NewCleanupService(repo repository.CleanupRepository) *CleanupService {
	return &CleanupService{
		repo: repo,
		log:  logger.Default().WithField("module", "cleanup"),
	}
}

// CleanTickets удаляет завершенные tickets и осиротевшие appointments
func (s *CleanupService) CleanTickets() error {
	s.log.Info("Начинаю очистку завершенных tickets и осиротевших appointments")

	// Получаем количество завершенных tickets
	ticketsCount, err := s.repo.GetTicketsCount()
	if err != nil {
		s.log.WithError(err).Error("Ошибка получения количества завершенных tickets")
		return err
	}

	// Получаем количество осиротевших appointments
	appointmentsCount, err := s.repo.GetOrphanedAppointmentsCount()
	if err != nil {
		s.log.WithError(err).Error("Ошибка получения количества осиротевших appointments")
		return err
	}

	s.log.WithFields(logrus.Fields{
		"completed_tickets_count":     ticketsCount,
		"orphaned_appointments_count": appointmentsCount,
	}).Info("Найдено записей для очистки")

	// Выполняем очистку
	if err := s.repo.TruncateTickets(); err != nil {
		s.log.WithError(err).Error("Ошибка очистки завершенных tickets и appointments")
		return err
	}

	s.log.Info("Очистка завершенных tickets и осиротевших appointments завершена успешно")
	return nil
}
