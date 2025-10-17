package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/pubsub"
	"ElectronicQueue/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"gorm.io/gorm"
)

// --- НОВЫЕ СТРУКТУРЫ ДЛЯ ПАРСИНГА ОТВЕТА 1С ---

// OneCSlot представляет один слот времени из JSON от 1С.
type OneCSlot struct {
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	IsAvailable bool   `json:"is_available"`
}

// OneCDoctor представляет одного врача из JSON от 1С.
type OneCDoctor struct {
	ID             uint       `json:"id"`
	FullName       string     `json:"full_name"`
	Specialization string     `json:"specialization"`
	Cabinet        any        `json:"cabinet"`
	Slots          []OneCSlot `json:"slots"`
}

// OneCScheduleResponse представляет корневую структуру JSON ответа от 1С.
type OneCScheduleResponse struct {
	Date    string       `json:"date"`
	MinTime string       `json:"min_start_time"`
	MaxTime string       `json:"max_end_time"`
	Doctors []OneCDoctor `json:"doctors"`
}

// --- Вспомогательная структура для слияния интервалов ---
type timeInterval struct {
	Start time.Time
	End   time.Time
}

// ScheduleService предоставляет методы для управления расписаниями.
type ScheduleService struct {
	scheduleRepo   repository.ScheduleRepository
	doctorRepo     repository.DoctorRepository
	oneCService    *OneCService
	broker         *pubsub.Broker
	log            *logger.AsyncLogger
	cachedSchedule json.RawMessage
	cacheLock      sync.RWMutex
}

// NewScheduleService создает новый экземпляр ScheduleService.
func NewScheduleService(
	scheduleRepo repository.ScheduleRepository,
	doctorRepo repository.DoctorRepository,
	oneCService *OneCService,
	broker *pubsub.Broker,
) *ScheduleService {
	return &ScheduleService{
		scheduleRepo:   scheduleRepo,
		doctorRepo:     doctorRepo,
		oneCService:    oneCService,
		broker:         broker,
		log:            logger.Default().WithField("module", "ScheduleService"),
		cachedSchedule: json.RawMessage("{}"),
	}
}

// StartPolling запускает фоновый процесс опроса 1С.
func (s *ScheduleService) StartPolling(ctx context.Context) {
	s.log.Info("Запуск периодического опроса расписания из 1С...")
	s.pollAggregateAndBroadcast()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.pollAggregateAndBroadcast()
		case <-ctx.Done():
			s.log.Info("Опрос расписания 1С остановлен.")
			return
		}
	}
}

// --- НОВАЯ ФУНКЦИЯ ДЛЯ СЛИЯНИЯ СЛОТОВ ---
func mergeIntervals(slots []OneCSlot, date string) []OneCSlot {
	if len(slots) < 2 {
		return slots
	}

	intervals := make(map[bool][]timeInterval)
	layout := "2006-01-02T15:04:05"

	for _, slot := range slots {
		start, err1 := time.Parse(layout, date+"T"+slot.StartTime)
		end, err2 := time.Parse(layout, date+"T"+slot.EndTime)
		if err1 != nil || err2 != nil {
			continue // Пропускаем некорректные слоты
		}
		intervals[slot.IsAvailable] = append(intervals[slot.IsAvailable], timeInterval{Start: start, End: end})
	}

	mergedSlots := []OneCSlot{}
	timeFormat := "15:04:05"

	for isAvailable, intervalList := range intervals {
		if len(intervalList) == 0 {
			continue
		}

		// Сортируем по времени начала
		sort.Slice(intervalList, func(i, j int) bool {
			return intervalList[i].Start.Before(intervalList[j].Start)
		})

		merged := []timeInterval{intervalList[0]}
		for i := 1; i < len(intervalList); i++ {
			last := &merged[len(merged)-1]
			current := intervalList[i]

			// Если текущий интервал пересекается с последним объединенным
			if current.Start.Before(last.End) || current.Start.Equal(last.End) {
				// Объединяем, выбирая максимальное время окончания
				if current.End.After(last.End) {
					last.End = current.End
				}
			} else {
				// Иначе добавляем новый интервал
				merged = append(merged, current)
			}
		}

		for _, interval := range merged {
			mergedSlots = append(mergedSlots, OneCSlot{
				StartTime:   interval.Start.Format(timeFormat),
				EndTime:     interval.End.Format(timeFormat),
				IsAvailable: isAvailable,
			})
		}
	}

	return mergedSlots
}

func (s *ScheduleService) pollAggregateAndBroadcast() {
	rawData, err := s.oneCService.GetDoctorSchedule()
	if err != nil {
		s.log.WithError(err).Error("Ошибка при опросе расписания из 1С")
		return
	}

	rawBytes, err := json.Marshal(rawData)
	if err != nil {
		s.log.WithError(err).Error("Ошибка сериализации сырого ответа от 1С")
		return
	}

	var oneCResp OneCScheduleResponse
	if err := json.Unmarshal(rawBytes, &oneCResp); err != nil {
		s.log.WithError(err).Error("Ошибка десериализации ответа от 1С в структуру OneCScheduleResponse")
		return
	}

	// Агрегация данных по врачам
	schedulesByDoctor := make(map[uint]DoctorScheduleModel)
	slotsByDoctor := make(map[uint][]OneCSlot)
	doctorInfo := make(map[uint]*OneCDoctor)

	for _, doctor := range oneCResp.Doctors {
		if doctor.ID == 0 {
			continue
		}
		if _, ok := doctorInfo[doctor.ID]; !ok {
			// Сохраняем информацию о враче (имя, кабинет и т.д.) при первом появлении
			newDoc := doctor
			doctorInfo[doctor.ID] = &newDoc
		}
		// Собираем ВСЕ слоты для каждого ID врача
		slotsByDoctor[doctor.ID] = append(slotsByDoctor[doctor.ID], doctor.Slots...)
	}

	// Обработка и слияние слотов для каждого врача
	for id, info := range doctorInfo {
		// --- ВЫЗОВ НОВОЙ ФУНКЦИИ СЛИЯНИЯ ---
		mergedDoctorSlots := mergeIntervals(slotsByDoctor[id], oneCResp.Date)

		docSchedule := DoctorScheduleModel{
			ID:             id,
			FullName:       info.FullName,
			Specialization: info.Specialization,
			Slots:          []TimeSlotModel{},
		}

		var cabinetPtr *int
		if cabinet, ok := info.Cabinet.(float64); ok {
			c := int(cabinet)
			cabinetPtr = &c
		}

		for _, slot := range mergedDoctorSlots {
			docSchedule.Slots = append(docSchedule.Slots, TimeSlotModel{
				StartTime:   slot.StartTime,
				EndTime:     slot.EndTime,
				IsAvailable: slot.IsAvailable,
				Cabinet:     cabinetPtr, // Кабинет одинаков для всех слотов одного врача
			})
		}
		schedulesByDoctor[id] = docSchedule
	}

	// Сортировка слотов внутри каждого врача
	for _, doctorSchedule := range schedulesByDoctor {
		sort.Slice(doctorSchedule.Slots, func(i, j int) bool {
			return doctorSchedule.Slots[i].StartTime < doctorSchedule.Slots[j].StartTime
		})
	}

	// Финальная сборка и сортировка списка врачей
	finalDoctorList := make([]DoctorScheduleModel, 0, len(schedulesByDoctor))
	for _, doc := range schedulesByDoctor {
		finalDoctorList = append(finalDoctorList, doc)
	}
	sort.Slice(finalDoctorList, func(i, j int) bool {
		return finalDoctorList[i].ID < finalDoctorList[j].ID
	})

	finalResponse := TodayScheduleResponse{
		Date:         oneCResp.Date,
		MinStartTime: oneCResp.MinTime,
		MaxEndTime:   oneCResp.MaxTime,
		Doctors:      finalDoctorList,
	}

	jsonBytes, err := json.Marshal(finalResponse)
	if err != nil {
		s.log.WithError(err).Error("Ошибка сериализации агрегированного расписания в JSON")
		return
	}

	s.cacheLock.Lock()
	s.cachedSchedule = jsonBytes
	s.cacheLock.Unlock()

	s.broker.Publish(string(jsonBytes))
	s.log.Info("Расписание из 1С успешно агрегировано, обновлено и разослано.")
}

// GetTodayScheduleState возвращает актуальное состояние расписания из кэша.
func (s *ScheduleService) GetTodayScheduleState() (json.RawMessage, error) {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()

	if len(s.cachedSchedule) <= 2 { // Проверка на пустой объект "{}"
		return nil, errors.New("кэш расписания пуст или не инициализирован")
	}

	return s.cachedSchedule, nil
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

// CreateSchedule создает новый слот в расписании (для админ-панели).
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

// DeleteSchedule удаляет слот из расписания по ID (для админ-панели).
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
