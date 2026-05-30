package services

import (
	"errors"
	"time"

	"projectvows/internal/dto"
	"projectvows/internal/models"
	"projectvows/internal/repositories"
)

// Check-in domain errors. Handlers map these to machine-readable codes.
var (
	ErrInvitationNotFound = errors.New("invitation_not_found")
	ErrNotAttending       = errors.New("not_attending")
	ErrInvalidEvent       = errors.New("invalid_event")
	ErrInvalidPax         = errors.New("invalid_pax")
	ErrAlreadyCheckedIn   = errors.New("already_checked_in")
)

type CheckinService struct {
	invRepo     *repositories.InvitationRepository
	checkinRepo *repositories.CheckinLogRepository
}

func NewCheckinService(invRepo *repositories.InvitationRepository, checkinRepo *repositories.CheckinLogRepository) *CheckinService {
	return &CheckinService{invRepo: invRepo, checkinRepo: checkinRepo}
}

// Lookup returns guest detail and per-event check-in status for a QR token.
func (s *CheckinService) Lookup(token string) (*dto.CheckinLookupResponse, error) {
	inv, err := s.invRepo.FindByQRToken(token)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	resp := &dto.CheckinLookupResponse{
		GuestName:       inv.GuestName,
		PaxCount:        inv.PaxCount,
		EventChoice:     inv.EventChoice,
		CheckedInEvents: []dto.CheckinEventStatus{},
	}

	// Build the list of events relevant to this guest's choice.
	for _, et := range eventsForChoice(inv.EventChoice) {
		status := dto.CheckinEventStatus{EventType: et}
		for i := range inv.CheckinLogs {
			if inv.CheckinLogs[i].EventType == et {
				t := inv.CheckinLogs[i].CheckedInAt
				status.CheckedInAt = &t
				break
			}
		}
		resp.CheckedInEvents = append(resp.CheckedInEvents, status)
	}

	return resp, nil
}

// Checkin validates and records a check-in. Returns the created log and guest info.
func (s *CheckinService) Checkin(req dto.CheckinRequest) (*models.Invitation, *models.CheckinLog, error) {
	inv, err := s.invRepo.FindByQRToken(req.QRCodeToken)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, nil, ErrInvitationNotFound
		}
		return nil, nil, err
	}

	// Must be attending.
	if inv.RSVPStatus != models.RSVPAttending {
		return nil, nil, ErrNotAttending
	}

	// event_choice must allow the requested event_type.
	if !choiceAllows(inv.EventChoice, req.EventType) {
		return nil, nil, ErrInvalidEvent
	}

	// actual_pax must be within the registered pax count.
	if inv.PaxCount == nil || req.ActualPax == 0 || req.ActualPax > *inv.PaxCount {
		return nil, nil, ErrInvalidPax
	}

	// Reject duplicates early for a clean message (the DB unique index is the
	// authoritative guard against races).
	exists, err := s.checkinRepo.ExistsForEvent(inv.ID, req.EventType)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, ErrAlreadyCheckedIn
	}

	log := &models.CheckinLog{
		InvitationID: inv.ID,
		EventType:    req.EventType,
		CheckedInPax: req.ActualPax,
		CheckedInAt:  time.Now(),
	}
	if req.ScannerName != "" {
		log.ScannerName = &req.ScannerName
	}

	if err := s.checkinRepo.Create(log); err != nil {
		if errors.Is(err, repositories.ErrDuplicateCheckin) {
			return nil, nil, ErrAlreadyCheckedIn
		}
		return nil, nil, err
	}

	return inv, log, nil
}

// eventsForChoice returns the event types relevant to a guest's event_choice.
func eventsForChoice(choice *string) []string {
	if choice == nil {
		return []string{}
	}
	switch *choice {
	case models.EventChoiceHolyMatrimony:
		return []string{models.CheckinEventHolyMatrimony}
	case models.EventChoiceReception:
		return []string{models.CheckinEventReception}
	case models.EventChoiceBoth:
		return []string{models.CheckinEventHolyMatrimony, models.CheckinEventReception}
	default:
		return []string{}
	}
}

// choiceAllows reports whether the guest's event_choice permits the event_type.
func choiceAllows(choice *string, eventType string) bool {
	if eventType != models.CheckinEventHolyMatrimony && eventType != models.CheckinEventReception {
		return false
	}
	for _, allowed := range eventsForChoice(choice) {
		if allowed == eventType {
			return true
		}
	}
	return false
}
