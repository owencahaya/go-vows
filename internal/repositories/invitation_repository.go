package repositories

import (
	"errors"

	"gorm.io/gorm"

	"projectvows/internal/dto"
	"projectvows/internal/models"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

// DB exposes the underlying *gorm.DB for services that need transactions.
func (r *InvitationRepository) DB() *gorm.DB { return r.db }

func (r *InvitationRepository) Create(inv *models.Invitation) error {
	return r.db.Create(inv).Error
}

func (r *InvitationRepository) Save(inv *models.Invitation) error {
	return r.db.Save(inv).Error
}

func (r *InvitationRepository) FindByID(id uint64) (*models.Invitation, error) {
	var inv models.Invitation
	err := r.db.Preload("Event").Preload("CheckinLogs").First(&inv, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvitationRepository) FindByIDs(ids []uint64) ([]models.Invitation, error) {
	var invs []models.Invitation
	err := r.db.Preload("Event").Where("id IN ?", ids).Find(&invs).Error
	return invs, err
}

func (r *InvitationRepository) FindByQRToken(token string) (*models.Invitation, error) {
	var inv models.Invitation
	err := r.db.Preload("Event").Preload("CheckinLogs").
		Where("qr_code_token = ?", token).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// List applies the optional filters and returns matching invitations.
func (r *InvitationRepository) List(f dto.InvitationFilter) ([]models.Invitation, error) {
	q := r.db.Preload("Event").Model(&models.Invitation{})

	if f.Tag != "" {
		q = q.Joins("JOIN events ON events.id = invitations.event_id").
			Where("events.tag = ?", f.Tag)
	}
	if f.EventID != 0 {
		q = q.Where("invitations.event_id = ?", f.EventID)
	}
	if f.RSVPStatus != "" {
		q = q.Where("invitations.rsvp_status = ?", f.RSVPStatus)
	}
	if f.InvitationStatus != "" {
		q = q.Where("invitations.invitation_status = ?", f.InvitationStatus)
	}
	if f.QRStatus != "" {
		q = q.Where("invitations.qr_status = ?", f.QRStatus)
	}

	var invs []models.Invitation
	err := q.Order("invitations.id DESC").Find(&invs).Error
	return invs, err
}

// FindPendingByTag returns invitations for an event with invitation_status = imported.
func (r *InvitationRepository) FindPendingByTag(tag string) ([]models.Invitation, error) {
	return r.findByTagAndConditions(tag, map[string]interface{}{
		"invitations.invitation_status": models.InvitationStatusImported,
	})
}

// FindUnansweredByTag returns invitations with rsvp_status = not_answered.
func (r *InvitationRepository) FindUnansweredByTag(tag string) ([]models.Invitation, error) {
	return r.findByTagAndConditions(tag, map[string]interface{}{
		"invitations.rsvp_status": models.RSVPNotAnswered,
	})
}

// FindAttendingByTag returns invitations with rsvp_status = attending.
func (r *InvitationRepository) FindAttendingByTag(tag string) ([]models.Invitation, error) {
	return r.findByTagAndConditions(tag, map[string]interface{}{
		"invitations.rsvp_status": models.RSVPAttending,
	})
}

// FindQRReadyByTag returns invitations eligible for QR generation/sending.
func (r *InvitationRepository) FindQRReadyByTag(tag string) ([]models.Invitation, error) {
	var invs []models.Invitation
	err := r.db.Preload("Event").
		Joins("JOIN events ON events.id = invitations.event_id").
		Where("events.tag = ?", tag).
		Where("invitations.rsvp_status = ?", models.RSVPAttending).
		Where("invitations.pax_count IS NOT NULL").
		Where("invitations.event_choice IS NOT NULL").
		Where("invitations.qr_sent_at IS NULL").
		Find(&invs).Error
	return invs, err
}

func (r *InvitationRepository) findByTagAndConditions(tag string, conds map[string]interface{}) ([]models.Invitation, error) {
	var invs []models.Invitation
	q := r.db.Preload("Event").
		Joins("JOIN events ON events.id = invitations.event_id").
		Where("events.tag = ?", tag)
	for k, v := range conds {
		q = q.Where(k+" = ?", v)
	}
	err := q.Find(&invs).Error
	return invs, err
}
