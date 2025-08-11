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

func (r *patientRepo) Create(patient *models.Patient) (*models.Patient, error) {
	err := r.db.Create(patient).Error
	if err != nil {
		return nil, err
	}
	return patient, nil
}

// Search ищет по ФИО, номеру ОМС и полному номеру паспорта.
func (r *patientRepo) Search(query string) ([]models.Patient, error) {
	var patients []models.Patient
	searchQuery := "%" + query + "%"

	err := r.db.Where(
		"full_name ILIKE ? OR oms_number LIKE ? OR (passport_series || passport_number) LIKE ?",
		searchQuery,
		searchQuery,
		searchQuery,
	).Limit(10).Find(&patients).Error

	return patients, err
}

func (r *patientRepo) FindByPassport(series, number string) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.Where("passport_series = ? AND passport_number = ?", series, number).First(&patient).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepo) FindByPhone(phone string) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.Where("regexp_replace(phone, '[^0-9]+', '', 'g') = ?", phone).First(&patient).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}
