package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/utils"
	"fmt"

	"gorm.io/gorm"
)

type AuthService struct {
	repo       repository.RegistrarRepository
	jwtManager *utils.JWTManager
}

func NewAuthService(repo repository.RegistrarRepository, jwtManager *utils.JWTManager) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) AuthenticateRegistrar(login, password string) (string, error) {
	registrar, err := s.repo.FindByLogin(login)
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

func (s *AuthService) CreateRegistrar(fullName, login, password string) (*models.Registrar, error) {
	// Проверяем, не занят ли логин
	_, err := s.repo.FindByLogin(login)
	if err == nil {
		return nil, fmt.Errorf("логин '%s' уже занят", login)
	}
	if err != gorm.ErrRecordNotFound {
		// Другая ошибка при проверке, например, проблема с БД
		return nil, err
	}

	// Хэшируем пароль
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("не удалось захэшировать пароль: %w", err)
	}

	newRegistrar := &models.Registrar{
		FullName:     fullName,
		Login:        login,
		PasswordHash: hashedPassword,
	}

	if err := s.repo.Create(newRegistrar); err != nil {
		return nil, fmt.Errorf("не удалось создать регистратора: %w", err)
	}

	return newRegistrar, nil
}
