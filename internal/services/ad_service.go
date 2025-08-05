package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"encoding/base64"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type AdService struct {
	repo repository.AdRepository
}

func NewAdService(repo repository.AdRepository) *AdService {
	return &AdService{repo: repo}
}

func (s *AdService) Create(req *models.CreateAdRequest) (*models.Ad, error) {
	picBytes, err := base64.StdEncoding.DecodeString(req.Picture)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 picture data: %w", err)
	}

	ad := &models.Ad{
		Picture:     picBytes,
		DurationSec: req.DurationSec,
		IsEnabled:   req.IsEnabled,
	}

	if err := s.repo.Create(ad); err != nil {
		return nil, fmt.Errorf("could not create ad: %w", err)
	}
	return ad, nil
}

func (s *AdService) GetAll() ([]models.Ad, error) {
	return s.repo.GetAll()
}

func (s *AdService) GetEnabled() ([]models.Ad, error) {
	return s.repo.GetEnabled()
}

func (s *AdService) GetByID(id uint) (*models.Ad, error) {
	return s.repo.GetByID(id)
}

func (s *AdService) Update(id uint, req *models.UpdateAdRequest) (*models.Ad, error) {
	ad, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ad with id %d not found", id)
		}
		return nil, err
	}

	if req.Picture != "" {
		picBytes, err := base64.StdEncoding.DecodeString(req.Picture)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 picture data: %w", err)
		}
		ad.Picture = picBytes
	}
	if req.DurationSec != nil {
		ad.DurationSec = *req.DurationSec
	}
	if req.IsEnabled != nil {
		ad.IsEnabled = *req.IsEnabled
	}

	if err := s.repo.Update(ad); err != nil {
		return nil, fmt.Errorf("could not update ad: %w", err)
	}
	return ad, nil
}

func (s *AdService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("could not delete ad: %w", err)
	}
	return nil
}
