package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/repository"
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

// CleanTickets выполняет полную очистку таблицы tickets.
func (s *CleanupService) CleanTickets() error {
	s.log.Info("Начинаю ежедневную очистку всех талонов и связанных записей.")

	// Получаем общее количество талонов перед удалением для логирования
	ticketsCount, err := s.repo.GetTotalTicketsCount()
	if err != nil {
		s.log.WithError(err).Error("Ошибка получения общего количества талонов.")
		// Продолжаем выполнение, даже если не удалось посчитать
	} else {
		s.log.WithField("total_tickets_to_delete", ticketsCount).Info("Найдено записей для полной очистки.")
	}

	// Выполняем полную очистку
	if err := s.repo.TruncateTickets(); err != nil {
		s.log.WithError(err).Error("Произошла ошибка во время полной очистки талонов.")
		return err
	}

	s.log.Info("Ежедневная очистка талонов и сброс счетчиков успешно завершены.")
	return nil
}
