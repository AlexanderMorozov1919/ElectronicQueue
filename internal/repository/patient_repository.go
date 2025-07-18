package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type patientRepo struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepo{db: db}
}

func (r *patientRepo) Create(patient *models.Patient) error {
	return r.db.Create(patient).Error
}

func (r *patientRepo) Update(patient *models.Patient) error {
	return r.db.Save(patient).Error
}

func (r *patientRepo) GetByID(id uint) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.First(&patient, id).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepo) FindByPassport(series, number string) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.Where("passport_series = ? AND passport_number = ?", series, number).First(&patient).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}
