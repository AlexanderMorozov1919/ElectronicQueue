package repository

import (
	"ElectronicQueue/internal/models/patient_model"

	"gorm.io/gorm"
)

type patientRepo struct {
	db *gorm.DB
}

// NewPatientRepository - конструктор для patientRepo.
func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepo{db: db}
}

func (r *patientRepo) Create(patient *patient_model.Patient) error {
	return r.db.Create(patient).Error
}

func (r *patientRepo) Update(patient *patient_model.Patient) error {
	return r.db.Save(patient).Error
}

func (r *patientRepo) GetByID(id uint) (*patient_model.Patient, error) {
	var patient patient_model.Patient
	if err := r.db.First(&patient, id).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepo) FindByPassport(series, number string) (*patient_model.Patient, error) {
	var patient patient_model.Patient
	if err := r.db.Where("passport_series = ? AND passport_number = ?", series, number).First(&patient).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}
