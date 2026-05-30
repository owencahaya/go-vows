package services

import (
	"time"

	"gorm.io/datatypes"

	"projectvows/internal/config"
	"projectvows/internal/models"
	"projectvows/internal/repositories"
)

// WebhookService handles inbound WhatsApp webhook events and drives the
// RSVP conversation flow. The actual Meta payload parsing is left as a TODO;
// the building-block methods below are ready to be wired up.
type WebhookService struct {
	cfg      *config.Config
	invRepo  *repositories.InvitationRepository
	logRepo  *repositories.WhatsappLogRepository
	whatsapp WhatsappService
	invSvc   *InvitationService
}

func NewWebhookService(
	cfg *config.Config,
	invRepo *repositories.InvitationRepository,
	logRepo *repositories.WhatsappLogRepository,
	whatsapp WhatsappService,
	invSvc *InvitationService,
) *WebhookService {
	return &WebhookService{
		cfg:      cfg,
		invRepo:  invRepo,
		logRepo:  logRepo,
		whatsapp: whatsapp,
		invSvc:   invSvc,
	}
}

// VerifyToken implements the Meta webhook verification handshake.
// Returns the challenge string and true when the verify token matches.
func (s *WebhookService) VerifyToken(mode, token, challenge string) (string, bool) {
	if mode == "subscribe" && token == s.cfg.MetaVerifyToken && s.cfg.MetaVerifyToken != "" {
		return challenge, true
	}
	return "", false
}

// HandleInbound processes a raw inbound webhook payload.
//
// TODO: Parse the actual Meta WhatsApp webhook structure:
//
//	entry[].changes[].value.messages[]  -> inbound messages
//	  - type == "interactive" -> button_reply / list_reply ids encode the answer
//	  - type == "text"        -> free text
//	entry[].changes[].value.statuses[]  -> delivery/read receipts
//	entry[].changes[].value.contacts[]  -> map wa_id to our whatsapp_number
//
// For now we store a generic inbound log if we can resolve an invitation by the
// sender's whatsapp number, otherwise we simply persist nothing and return nil.
func (s *WebhookService) HandleInbound(rawPayload []byte) error {
	// TODO: extract sender wa_id + interactive reply id from rawPayload and
	// route to the appropriate Update* method below. For now we record the raw
	// payload as a generic "other" inbound log against invitation 0 is not
	// possible (FK), so we no-op until parsing is implemented.
	_ = rawPayload
	return nil
}

// logInbound writes a generic inbound whatsapp_logs row.
func (s *WebhookService) logInbound(invitationID uint64, msgType string, payload datatypes.JSON, whatsappMessageID *string) error {
	now := time.Now()
	return s.logRepo.Create(&models.WhatsappLog{
		InvitationID:      invitationID,
		MessageType:       msgType,
		Direction:         models.DirectionInbound,
		WhatsappMessageID: whatsappMessageID,
		Status:            models.WALogReceived,
		Payload:           payload,
		ReceivedAt:        &now,
	})
}

// --- Conversation state mutators (ready for the webhook parser to call) ---

// UpdateRSVPStatus sets the RSVP answer for an invitation.
func (s *WebhookService) UpdateRSVPStatus(inv *models.Invitation, status string) error {
	inv.RSVPStatus = status
	return s.invRepo.Save(inv)
}

// UpdatePaxCount sets how many people will attend.
func (s *WebhookService) UpdatePaxCount(inv *models.Invitation, pax uint16) error {
	inv.PaxCount = &pax
	return s.invRepo.Save(inv)
}

// UpdateEventChoice sets which events the guest will attend.
func (s *WebhookService) UpdateEventChoice(inv *models.Invitation, choice string) error {
	inv.EventChoice = &choice
	return s.invRepo.Save(inv)
}

// UpdateGiftInterest records whether the guest wants gift/bank info.
func (s *WebhookService) UpdateGiftInterest(inv *models.Invitation, interest string) error {
	inv.GiftInterest = interest
	return s.invRepo.Save(inv)
}

// TriggerNextMessage decides and sends the next message in the RSVP flow.
//
// TODO: implement the real conversation state machine, e.g.:
//
//	not_answered            -> (no-op, waiting)
//	attending + no pax      -> ask_pax
//	attending + pax + no choice -> ask_event_choice
//	attending + complete    -> TriggerQRGeneration
func (s *WebhookService) TriggerNextMessage(inv *models.Invitation) error {
	// TODO: send the contextually-appropriate interactive message.
	_ = inv
	return nil
}

// TriggerQRGeneration generates and sends the QR once RSVP is complete.
func (s *WebhookService) TriggerQRGeneration(inv *models.Invitation) error {
	if inv.RSVPStatus != models.RSVPAttending || inv.PaxCount == nil || inv.EventChoice == nil {
		return nil // not ready yet
	}
	// Reuse the invitation service QR pipeline for a single invitation.
	if inv.Event == nil {
		return nil
	}
	_, err := s.invSvc.ResendQR(inv.Event.Tag, []uint64{inv.ID})
	return err
}
