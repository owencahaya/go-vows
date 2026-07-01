package services

import (
	"net/url"
	"strings"
	"time"

	twilioClient "github.com/twilio/twilio-go/client"
	"gorm.io/datatypes"

	"projectvows/internal/config"
	"projectvows/internal/models"
	"projectvows/internal/repositories"
)

// WebhookService handles inbound Twilio WhatsApp webhooks: delivery/status
// callbacks and inbound guest messages. Twilio posts application/x-www-form-
// urlencoded payloads and authenticates them with an X-Twilio-Signature header
// (there is no Meta-style GET verification handshake).
type WebhookService struct {
	cfg       *config.Config
	invRepo   *repositories.InvitationRepository
	logRepo   *repositories.WhatsappLogRepository
	whatsapp  WhatsappService
	invSvc    *InvitationService
	validator twilioClient.RequestValidator
}

func NewWebhookService(
	cfg *config.Config,
	invRepo *repositories.InvitationRepository,
	logRepo *repositories.WhatsappLogRepository,
	whatsapp WhatsappService,
	invSvc *InvitationService,
) *WebhookService {
	return &WebhookService{
		cfg:       cfg,
		invRepo:   invRepo,
		logRepo:   logRepo,
		whatsapp:  whatsapp,
		invSvc:    invSvc,
		validator: twilioClient.NewRequestValidator(cfg.TwilioAuthToken),
	}
}

// ValidateSignature verifies the X-Twilio-Signature header for a POST webhook.
// When no auth token is configured (e.g. local dev / sandbox), validation is
// skipped and the request is accepted.
func (s *WebhookService) ValidateSignature(fullURL string, form url.Values, signature string) bool {
	if s.cfg.TwilioAuthToken == "" {
		return true
	}
	params := make(map[string]string, len(form))
	for k := range form {
		params[k] = form.Get(k)
	}
	return s.validator.Validate(fullURL, params, signature)
}

// HandleTwilioCallback routes a parsed Twilio webhook payload. A payload with a
// MessageStatus field is a delivery/status callback; otherwise it is treated as
// an inbound message from a guest.
func (s *WebhookService) HandleTwilioCallback(form url.Values) error {
	if status := form.Get("MessageStatus"); status != "" {
		return s.handleStatusCallback(form.Get("MessageSid"), status)
	}
	return s.handleInboundMessage(form)
}

// handleStatusCallback maps Twilio's message status onto our whatsapp_logs
// status column for the matching outbound log.
func (s *WebhookService) handleStatusCallback(messageSid, twilioStatus string) error {
	if messageSid == "" {
		return nil
	}
	return s.logRepo.UpdateStatusByMessageID(messageSid, mapTwilioStatus(twilioStatus))
}

// handleInboundMessage resolves the sender and records an inbound log. The RSVP
// conversation state machine (UpdateRSVPStatus etc.) can be driven from here.
func (s *WebhookService) handleInboundMessage(form url.Values) error {
	from := strings.TrimPrefix(form.Get("From"), "whatsapp:")
	if from == "" {
		return nil
	}
	inv, err := s.invRepo.FindByWhatsappNumber(from)
	if err != nil {
		// Unknown sender: nothing to attach the log to (invitation_id is a FK).
		return nil
	}

	var messageID *string
	if sid := form.Get("MessageSid"); sid != "" {
		messageID = &sid
	}
	payload := toJSON(formToMap(form))
	return s.logInbound(inv.ID, models.MsgTypeOther, payload, messageID)
}

// mapTwilioStatus translates a Twilio message status to our log status enum.
func mapTwilioStatus(twilioStatus string) string {
	switch strings.ToLower(twilioStatus) {
	case "failed", "undelivered":
		return models.WALogFailed
	case "sent", "delivered", "read":
		return models.WALogSent
	default: // queued, accepted, scheduled, sending, ...
		return models.WALogPending
	}
}

func formToMap(form url.Values) map[string]string {
	m := make(map[string]string, len(form))
	for k := range form {
		m[k] = form.Get(k)
	}
	return m
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

// TriggerQRGeneration generates and sends the QR once RSVP is complete.
func (s *WebhookService) TriggerQRGeneration(inv *models.Invitation) error {
	if inv.RSVPStatus != models.RSVPAttending || inv.PaxCount == nil || inv.EventChoice == nil {
		return nil // not ready yet
	}
	if inv.Event == nil {
		return nil
	}
	_, err := s.invSvc.ResendQR(inv.Event.Tag, []uint64{inv.ID})
	return err
}
