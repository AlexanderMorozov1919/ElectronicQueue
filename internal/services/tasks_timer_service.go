package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/logger"
)

type TasksTimerService struct {
	cleanupService *CleanupService
	config         *config.Config
	log            *logger.AsyncLogger
}

func NewTasksTimerService(cleanupService *CleanupService, config *config.Config) *TasksTimerService {
	return &TasksTimerService{
		cleanupService: cleanupService,
		config:         config,
		log:            logger.Default().WithField("module", "tasks_timer"),
	}
}

// Start запускает планировщик задач
func (s *TasksTimerService) Start(ctx context.Context) {
	s.log.Info("Планировщик задач запущен")

	for {
		// Вычисляем время следующего запуска
		nextRun := s.calculateNextRun()
		s.log.WithField("next_run", nextRun.Format("2006-01-02 15:04:05")).Info("Следующий запуск очистки: " + nextRun.Format("2006-01-02 15:04:05"))

		// Ждем до времени выполнения
		select {
		case <-time.After(time.Until(nextRun)):
			// Выполняем очистку
			if err := s.cleanupService.CleanTickets(); err != nil {
				s.log.WithError(err).Error("Ошибка выполнения очистки tickets")
			}
		case <-ctx.Done():
			s.log.Info("Планировщик задач остановлен")
			return
		}
	}
}

// calculateNextRun вычисляет время следующего запуска
func (s *TasksTimerService) calculateNextRun() time.Time {
	now := time.Now()

	// Парсим время из конфига (формат "HH:MM")
	timeParts := strings.Split(s.config.MaintenanceTime, ":")
	if len(timeParts) != 2 {
		s.log.WithField("maintenance_time", s.config.MaintenanceTime).Error("Неверный формат времени в конфиге")
		// Используем 00:00 по умолчанию
		timeParts = []string{"00", "00"}
	}

	hour := 0
	minute := 0
	fmt.Sscanf(timeParts[0], "%d", &hour)
	fmt.Sscanf(timeParts[1], "%d", &minute)

	// Вычисляем завтра в указанное время
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	// Если время уже прошло сегодня, то следующий запуск завтра
	if next.Before(now) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}
