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
	if (req.Picture == "" && req.Video == "") || (req.Picture != "" && req.Video != "") {
		return nil, errors.New("you must provide either a picture or a video, but not both")
	}

	ad := &models.Ad{
		IsEnabled:   req.IsEnabled,
		ReceptionOn: req.ReceptionOn,
		ScheduleOn:  req.ScheduleOn,
	}

	if req.Picture != "" {
		picBytes, err := base64.StdEncoding.DecodeString(req.Picture)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 picture data: %w", err)
		}
		ad.Picture = picBytes
		ad.Video = nil

		if req.DurationSec != nil {
			ad.DurationSec = *req.DurationSec
		} else {
			ad.DurationSec = 5 // Default for images
		}
		ad.RepeatCount = 1 // Default for images
	} else { // Video is provided
		videoBytes, err := base64.StdEncoding.DecodeString(req.Video)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 video data: %w", err)
		}
		ad.Video = videoBytes
		ad.Picture = nil

		ad.DurationSec = 5 // Default value since DB field is NOT NULL
		if req.RepeatCount != nil {
			ad.RepeatCount = *req.RepeatCount
		} else {
			ad.RepeatCount = 1 // Default for videos
		}
	}

	if err := s.repo.Create(ad); err != nil {
		return nil, fmt.Errorf("could not create ad: %w", err)
	}
	return ad, nil
}

func (s *AdService) GetAll() ([]models.Ad, error) {
	return s.repo.GetAll()
}

func (s *AdService) GetEnabled(screen string) ([]models.Ad, error) {
	if screen != "reception" && screen != "schedule" {
		return nil, fmt.Errorf("invalid screen type provided: %s", screen)
	}
	return s.repo.GetEnabledFor(screen)
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
		ad.Video = nil // Clear video if picture is updated
	} else if req.Video != "" {
		videoBytes, err := base64.StdEncoding.DecodeString(req.Video)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 video data: %w", err)
		}
		ad.Video = videoBytes
		ad.Picture = nil // Clear picture if video is updated
	}

	if req.DurationSec != nil {
		ad.DurationSec = *req.DurationSec
	}
	if req.RepeatCount != nil {
		ad.RepeatCount = *req.RepeatCount
	}
	if req.IsEnabled != nil {
		ad.IsEnabled = *req.IsEnabled
	}
	if req.ReceptionOn != nil {
		ad.ReceptionOn = *req.ReceptionOn
	}
	if req.ScheduleOn != nil {
		ad.ScheduleOn = *req.ScheduleOn
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
