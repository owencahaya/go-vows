package repositories

import (
	"errors"

	"gorm.io/gorm"

	"projectvows/internal/models"
)

// ErrDuplicateCheckin is returned when an invitation is already checked in for an event.
var ErrDuplicateCheckin = errors.New("already checked in")

type CheckinLogRepository struct {
	db *gorm.DB
}

func NewCheckinLogRepository(db *gorm.DB) *CheckinLogRepository {
	return &CheckinLogRepository{db: db}
}

// Create inserts a check-in log. Returns ErrDuplicateCheckin when the
// invitation_id + event_type unique constraint is violated.
func (r *CheckinLogRepository) Create(log *models.CheckinLog) error {
	err := r.db.Create(log).Error
	if err != nil && errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateCheckin
	}
	return err
}

// ExistsForEvent reports whether a check-in already exists for the invitation/event.
func (r *CheckinLogRepository) ExistsForEvent(invitationID uint64, eventType string) (bool, error) {
	var count int64
	err := r.db.Model(&models.CheckinLog{}).
		Where("invitation_id = ? AND event_type = ?", invitationID, eventType).
		Count(&count).Error
	return count > 0, err
}
