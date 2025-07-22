package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/utils"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	registrarRepo repository.RegistrarRepository
	doctorRepo    repository.DoctorRepository
	jwtManager    *utils.JWTManager
}

func NewAuthService(registrarRepo repository.RegistrarRepository, doctorRepo repository.DoctorRepository, jwtManager *utils.JWTManager) *AuthService {
	return &AuthService{
		registrarRepo: registrarRepo,
		doctorRepo:    doctorRepo,
		jwtManager:    jwtManager,
	}
}

func (s *AuthService) AuthenticateRegistrar(login, password string) (string, error) {
	registrar, err := s.registrarRepo.FindByLogin(login)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("неверный логин или пароль")
		}
		return "", err
	}

	if !utils.CheckPasswordHash(password, registrar.PasswordHash) {
		return "", fmt.Errorf("неверный логин или пароль")
	}

	return s.jwtManager.GenerateJWT(registrar.RegistrarID, "registrar")
}

func (s *AuthService) AuthenticateDoctor(login, password string) (string, *models.Doctor, error) {
	doctor, err := s.doctorRepo.FindByLogin(login)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, fmt.Errorf("неверный логин или пароль")
		}
		return "", nil, err
	}

	if !utils.CheckPasswordHash(password, doctor.PasswordHash) {
		return "", nil, fmt.Errorf("неверный логин или пароль")
	}

	token, err := s.jwtManager.GenerateJWT(doctor.ID, "doctor")
	if err != nil {
		return "", nil, err
	}

	return token, doctor, nil
}

func (s *AuthService) CreateRegistrar(windowNumber int, login, password string) (*models.Registrar, error) {
	_, err := s.registrarRepo.FindByLogin(login)
	if err == nil {
		return nil, fmt.Errorf("логин '%s' уже занят", login)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("не удалось захэшировать пароль: %w", err)
	}

	newRegistrar := &models.Registrar{
		WindowNumber: windowNumber,
		Login:        login,
		PasswordHash: string(hashedPassword),
	}

	if err := s.registrarRepo.Create(newRegistrar); err != nil {
		return nil, fmt.Errorf("не удалось создать регистратора: %w", err)
	}

	return newRegistrar, nil
}

func (s *AuthService) CreateDoctor(fullName, specialization, login, password string) (*models.Doctor, error) {
	// Проверяем, не занят ли логин
	_, err := s.doctorRepo.FindByLogin(login)
	if err == nil {
		return nil, fmt.Errorf("логин '%s' уже занят", login)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("не удалось захэшировать пароль: %w", err)
	}

	newDoctor := &models.Doctor{
		FullName:       fullName,
		Specialization: specialization,
		Login:          login,
		PasswordHash:   hashedPassword,
	}

	if err := s.doctorRepo.Create(newDoctor); err != nil {
		return nil, fmt.Errorf("не удалось создать врача: %w", err)
	}

	return newDoctor, nil
}