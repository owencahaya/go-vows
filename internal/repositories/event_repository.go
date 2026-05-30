package repositories

import (
	"errors"

	"gorm.io/gorm"

	"projectvows/internal/models"
)

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("record not found")

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *EventRepository) List() ([]models.Event, error) {
	var events []models.Event
	err := r.db.Order("id DESC").Find(&events).Error
	return events, err
}

func (r *EventRepository) FindByID(id uint64) (*models.Event, error) {
	var event models.Event
	err := r.db.First(&event, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) FindByTag(tag string) (*models.Event, error) {
	var event models.Event
	err := r.db.Where("tag = ?", tag).First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}
