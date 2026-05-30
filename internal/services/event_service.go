package services

import (
	"projectvows/internal/dto"
	"projectvows/internal/models"
	"projectvows/internal/repositories"
)

type EventService struct {
	repo *repositories.EventRepository
}

func NewEventService(repo *repositories.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) Create(req dto.CreateEventRequest) (*models.Event, error) {
	event := &models.Event{
		Tag:                   req.Tag,
		CoupleName:            req.CoupleName,
		HolyMatrimonyDate:     req.HolyMatrimonyDate,
		HolyMatrimonyLocation: req.HolyMatrimonyLocation,
		ReceptionDate:         req.ReceptionDate,
		ReceptionLocation:     req.ReceptionLocation,
		GiftAddress:           req.GiftAddress,
		BankAccount:           req.BankAccount,
	}
	if err := s.repo.Create(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) List() ([]models.Event, error) {
	return s.repo.List()
}

func (s *EventService) GetByID(id uint64) (*models.Event, error) {
	return s.repo.FindByID(id)
}
