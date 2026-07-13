package services

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"

	"projectvows/internal/config"
	"projectvows/internal/models"
	"projectvows/internal/repositories"
)

// WebhookService handles inbound Meta WhatsApp Cloud API webhooks: the GET
// subscription verification handshake, and POST payloads carrying inbound
// guest messages and delivery/status updates.
// https://developers.facebook.com/docs/whatsapp/cloud-api/guides/set-up-webhooks
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

// VerifyToken implements the Meta webhook subscription handshake:
// GET /api/webhook/whatsapp?hub.mode=subscribe&hub.verify_token=...&hub.challenge=...
// Returns the challenge string and true when mode is "subscribe" and the
// token matches META_VERIFY_TOKEN.
func (s *WebhookService) VerifyToken(mode, token, challenge string) (string, bool) {
	if mode == "subscribe" && s.cfg.MetaVerifyToken != "" && token == s.cfg.MetaVerifyToken {
		return challenge, true
	}
	return "", false
}

// metaWebhookPayload mirrors the Meta WhatsApp Cloud API webhook shape:
// entry[].changes[].value.{messages[],statuses[]}.
type metaWebhookPayload struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []metaInboundMessage `json:"messages"`
				Statuses []metaStatusUpdate   `json:"statuses"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

type metaInboundMessage struct {
	From string `json:"from"`
	ID   string `json:"id"`
	Type string `json:"type"`
	Text struct {
		Body string `json:"body"`
	} `json:"text"`
}

type metaStatusUpdate struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// HandleInbound parses a raw Meta webhook POST body and processes every
// inbound message and status update it contains.
func (s *WebhookService) HandleInbound(rawPayload []byte) error {
	var payload metaWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return err
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if err := s.handleInboundMessage(msg, rawPayload); err != nil {
					return err
				}
			}
			for _, status := range change.Value.Statuses {
				if err := s.handleStatusUpdate(status); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// handleInboundMessage resolves the sender and records an inbound log. The
// RSVP conversation state machine (UpdateRSVPStatus etc.) can be driven from here.
func (s *WebhookService) handleInboundMessage(msg metaInboundMessage, rawPayload []byte) error {
	if msg.From == "" {
		return nil
	}
	inv, err := s.invRepo.FindByWhatsappNumber(msg.From)
	if err != nil {
		// Unknown sender: nothing to attach the log to (invitation_id is a FK).
		return nil
	}

	var messageID *string
	if msg.ID != "" {
		messageID = &msg.ID
	}
	return s.logInbound(inv.ID, models.MsgTypeOther, datatypes.JSON(rawPayload), messageID)
}

// handleStatusUpdate maps Meta's message status onto our whatsapp_logs status
// column for the matching outbound log.
func (s *WebhookService) handleStatusUpdate(status metaStatusUpdate) error {
	if status.ID == "" {
		return nil
	}
	return s.logRepo.UpdateStatusByMessageID(status.ID, mapMetaStatus(status.Status))
}

// mapMetaStatus translates a Meta message status to our log status enum.
func mapMetaStatus(metaStatus string) string {
	switch metaStatus {
	case "failed":
		return models.WALogFailed
	case "sent", "delivered", "read":
		return models.WALogSent
	default:
		return models.WALogPending
	}
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
