package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
)

type RegistrarService struct {
	priorityRepo repository.RegistrarPriorityRepository
	serviceRepo  repository.ServiceRepository
}

func NewRegistrarService(priorityRepo repository.RegistrarPriorityRepository, serviceRepo repository.ServiceRepository) *RegistrarService {
	return &RegistrarService{priorityRepo: priorityRepo, serviceRepo: serviceRepo}
}

func (s *RegistrarService) GetPriorities(registrarID uint) ([]models.Service, error) {
	return s.priorityRepo.GetPriorities(registrarID)
}

func (s *RegistrarService) SetPriorities(registrarID uint, serviceIDs []uint) error {
	return s.priorityRepo.SetPriorities(registrarID, serviceIDs)
}

func (s *RegistrarService) GetAllServices() ([]models.Service, error) {
	return s.serviceRepo.GetAll()
}
