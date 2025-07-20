package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
)

type PatientService struct {
	repo repository.PatientRepository
}

func NewPatientService(repo repository.PatientRepository) *PatientService {
	return &PatientService{repo: repo}
}

func (s *PatientService) CreatePatient(req *models.CreatePatientRequest) (*models.Patient, error) {
	patient := &models.Patient{
		PassportSeries: req.PassportSeries,
		PassportNumber: req.PassportNumber,
		FullName:       req.FullName,
		BirthDate:      req.BirthDate,
		Phone:          req.Phone,
		OmsNumber:      req.OmsNumber,
	}
	createdPatient, err := s.repo.Create(patient)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пациента в репозитории: %w", err)
	}
	return createdPatient, nil
}

// SearchPatients вызывает универсальный поиск.
func (s *PatientService) SearchPatients(query string) ([]models.Patient, error) {
	if len(query) < 2 {
		return []models.Patient{}, nil
	}
	patients, err := s.repo.Search(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска пациентов в репозитории: %w", err)
	}
	return patients, nil
}
