package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type doctorRepo struct {
	db *gorm.DB
}

func NewDoctorRepository(db *gorm.DB) DoctorRepository {
	return &doctorRepo{db: db}
}

func (r *doctorRepo) Create(doctor *models.Doctor) error {
	return r.db.Create(doctor).Error
}

func (r *doctorRepo) Update(doctor *models.Doctor) error {
	return r.db.Save(doctor).Error
}

func (r *doctorRepo) Delete(id uint) error {
	return r.db.Delete(&models.Doctor{}, id).Error
}

func (r *doctorRepo) GetByID(id uint) (*models.Doctor, error) {
	var doctor models.Doctor
	if err := r.db.First(&doctor, id).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *doctorRepo) GetAll(onlyActive bool) ([]models.Doctor, error) {
	var doctors []models.Doctor
	query := r.db
	if onlyActive {
		query = query.Where("status = ?", models.DoctorStatusActive)
	}
	if err := query.Find(&doctors).Error; err != nil {
		return nil, err
	}
	return doctors, nil
}

// GetAnyDoctor возвращает первого активного врача, найденного в базе данных.
func (r *doctorRepo) GetAnyDoctor() (*models.Doctor, error) {
	var doctor models.Doctor
	if err := r.db.Where("status = ?", models.DoctorStatusActive).Order("doctor_id asc").First(&doctor).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *doctorRepo) FindByLogin(login string) (*models.Doctor, error) {
	var doctor models.Doctor
	if err := r.db.Where("login = ?", login).First(&doctor).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}

// UpdateStatus обновляет статус врача
func (r *doctorRepo) UpdateStatus(doctorID uint, status models.DoctorStatus) error {
	return r.db.Model(&models.Doctor{}).Where("doctor_id = ?", doctorID).Update("status", status).Error
}
