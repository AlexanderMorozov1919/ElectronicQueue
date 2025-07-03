package repository

import (
	"github.com/AlexanderMorozov1919/ElectronicQueue/internal/models/doctor_model"
	"gorm.io/gorm"
)

type doctorRepo struct {
	db *gorm.DB
}

// NewDoctorRepository - конструктор для doctorRepo.
func NewDoctorRepository(db *gorm.DB) DoctorRepository {
	return &doctorRepo{db: db}
}

func (r *doctorRepo) Create(doctor *doctor_model.Doctor) error {
	return r.db.Create(doctor).Error
}

func (r *doctorRepo) Update(doctor *doctor_model.Doctor) error {
	return r.db.Save(doctor).Error
}

func (r *doctorRepo) Delete(id uint) error {
	return r.db.Delete(&doctor_model.Doctor{}, id).Error
}

func (r *doctorRepo) GetByID(id uint) (*doctor_model.Doctor, error) {
	var doctor doctor_model.Doctor
	if err := r.db.First(&doctor, id).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *doctorRepo) GetAll(onlyActive bool) ([]doctor_model.Doctor, error) {
	var doctors []doctor_model.Doctor
	query := r.db
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&doctors).Error; err != nil {
		return nil, err
	}
	return doctors, nil
}
