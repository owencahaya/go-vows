package services

import (
	"fmt"

	"projectvows/internal/config"
	"projectvows/internal/models"
)

// SendResult is the normalized outcome of a WhatsApp send operation.
type SendResult struct {
	WhatsappMessageID string                 // Meta message id (wamid...)
	Response          map[string]interface{} // raw Meta response (for logging)
	Err               error                  // non-nil on failure
}

// WhatsappService abstracts the Meta WhatsApp Cloud API. The concrete
// implementation here is a stub; swap it for a real HTTP client later.
type WhatsappService interface {
	// SendInvitation sends the initial/resend invitation message.
	SendInvitation(inv *models.Invitation, messageType string) SendResult
	// SendReminder sends a reminder to an attending guest.
	SendReminder(inv *models.Invitation) SendResult
	// SendQR sends a previously-uploaded QR media to the guest.
	SendQR(inv *models.Invitation, mediaID string) SendResult
	// SendText sends a plain text message (used by webhook conversation flow).
	SendText(inv *models.Invitation, body string) SendResult
}

type whatsappServiceStub struct {
	cfg *config.Config
}

// NewWhatsappService returns the stub implementation.
func NewWhatsappService(cfg *config.Config) WhatsappService {
	return &whatsappServiceStub{cfg: cfg}
}

func (s *whatsappServiceStub) SendInvitation(inv *models.Invitation, messageType string) SendResult {
	// TODO: Replace with a real call to the Meta Cloud API:
	//   POST https://graph.facebook.com/{API_VERSION}/{PHONE_NUMBER_ID}/messages
	//   Authorization: Bearer {META_ACCESS_TOKEN}
	//   Body: a template or interactive message addressed to inv.WhatsappNumber.
	return s.fakeSend(inv, fmt.Sprintf("invitation:%s", messageType))
}

func (s *whatsappServiceStub) SendReminder(inv *models.Invitation) SendResult {
	// TODO: send reminder template message via Meta Cloud API.
	return s.fakeSend(inv, "reminder")
}

func (s *whatsappServiceStub) SendQR(inv *models.Invitation, mediaID string) SendResult {
	// TODO: send an image message referencing the uploaded media id via Meta Cloud API.
	return s.fakeSend(inv, fmt.Sprintf("qr:%s", mediaID))
}

func (s *whatsappServiceStub) SendText(inv *models.Invitation, body string) SendResult {
	// TODO: send a free-form text message via Meta Cloud API.
	return s.fakeSend(inv, "text")
}

// fakeSend returns a deterministic fake success result so the rest of the
// pipeline (logging, status updates) can be exercised end-to-end.
func (s *whatsappServiceStub) fakeSend(inv *models.Invitation, kind string) SendResult {
	msgID := fmt.Sprintf("wamid.STUB-%d-%s", inv.ID, kind)
	return SendResult{
		WhatsappMessageID: msgID,
		Response: map[string]interface{}{
			"stub":     true,
			"kind":     kind,
			"to":       inv.WhatsappNumber,
			"messages": []map[string]string{{"id": msgID}},
		},
	}
}
