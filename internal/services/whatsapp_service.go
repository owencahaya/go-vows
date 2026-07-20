package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"projectvows/internal/config"
	"projectvows/internal/models"
)

// SendResult is the normalized outcome of a WhatsApp send operation.
// Field names are preserved from the previous integration so callers and the
// whatsapp_logs writer stay unchanged; the values now come from Meta.
type SendResult struct {
	WhatsappMessageID string                 // Meta message id (wamid...)
	Response          map[string]interface{} // raw Meta response (for logging)
	Err               error                  // non-nil on failure
}

// WhatsappService abstracts the WhatsApp provider (now Meta WhatsApp Cloud
// API). The interface is unchanged from the previous Twilio-based abstraction
// so callers (InvitationService, WebhookService) require no changes.
type WhatsappService interface {
	// SendInvitation sends the initial/resend invitation message.
	//
	// When META_TEMPLATE_NAME_INVITATION is configured, it sends that approved
	// template with imageURL as the required header image and the guest name
	// as the "guest_name" named body parameter (imageURL is then required —
	// the template's header component is IMAGE type). When no template is
	// configured, imageURL is optional: sent as an image message (Meta "link"
	// media) with the invitation text as caption, or a plain text message
	// when empty.
	SendInvitation(inv *models.Invitation, messageType, imageURL string) SendResult
	// SendReminder sends a reminder to an attending guest.
	SendReminder(inv *models.Invitation) SendResult
	// SendQR sends the QR image. mediaURL is a publicly reachable HTTPS URL
	// (the existing stateless /qr endpoint) that Meta fetches as image.link.
	SendQR(inv *models.Invitation, mediaURL string) SendResult
	// SendText sends a plain text message (used by the webhook conversation flow).
	SendText(inv *models.Invitation, body string) SendResult
}

type metaWhatsappService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewWhatsappService returns the Meta-backed implementation. The constructor
// signature is unchanged so the DI wiring in main.go stays the same.
func NewWhatsappService(cfg *config.Config) WhatsappService {
	return &metaWhatsappService{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *metaWhatsappService) SendInvitation(inv *models.Invitation, messageType, imageURL string) SendResult {
	if s.cfg.MetaTemplateNameInvitation != "" {
		if s.cfg.MetaTemplateLanguage == "" {
			return SendResult{Err: fmt.Errorf("META_TEMPLATE_LANGUAGE is not configured for template %q", s.cfg.MetaTemplateNameInvitation)}
		}
		if imageURL == "" {
			return SendResult{Err: fmt.Errorf("template %q requires an image_url (header image)", s.cfg.MetaTemplateNameInvitation)}
		}
		return s.sendInvitationTemplate(inv, imageURL)
	}

	body := fmt.Sprintf(
		"Halo %s, Anda diundang ke acara pernikahan %s. Mohon konfirmasi kehadiran Anda.",
		inv.GuestName, coupleName(inv),
	)
	if imageURL != "" {
		return s.sendImage(inv, imageURL, body)
	}
	return s.sendText(inv, body)
}

func (s *metaWhatsappService) SendReminder(inv *models.Invitation) SendResult {
	body := fmt.Sprintf(
		"Halo %s, ini pengingat untuk acara pernikahan %s. Sampai jumpa!",
		inv.GuestName, coupleName(inv),
	)
	return s.sendText(inv, body)
}

func (s *metaWhatsappService) SendQR(inv *models.Invitation, mediaURL string) SendResult {
	body := fmt.Sprintf(
		"Halo %s, berikut QR code untuk check-in di acara %s.",
		inv.GuestName, coupleName(inv),
	)
	return s.sendImage(inv, mediaURL, body)
}

func (s *metaWhatsappService) SendText(inv *models.Invitation, body string) SendResult {
	return s.sendText(inv, body)
}

// sendText sends a freeform text message via the Meta Cloud API.
// https://developers.facebook.com/docs/whatsapp/cloud-api/messages/text-messages
func (s *metaWhatsappService) sendText(inv *models.Invitation, body string) SendResult {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                normalizeIndonesianPhone(inv.WhatsappNumber),
		"type":              "text",
		"text": map[string]interface{}{
			"preview_url": false,
			"body":        body,
		},
	}
	return s.send(payload)
}

// sendImage sends an image message referenced by a public HTTPS link, with an
// optional caption. https://developers.facebook.com/docs/whatsapp/cloud-api/messages/image-messages
func (s *metaWhatsappService) sendImage(inv *models.Invitation, link, caption string) SendResult {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                normalizeIndonesianPhone(inv.WhatsappNumber),
		"type":              "image",
		"image": map[string]interface{}{
			"link":    link,
			"caption": caption,
		},
	}
	return s.send(payload)
}

// sendInvitationTemplate sends the configured invitation template with an
// IMAGE header and a "guest_name" named body parameter.
// https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-message-templates
func (s *metaWhatsappService) sendInvitationTemplate(inv *models.Invitation, headerImageURL string) SendResult {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                normalizeIndonesianPhone(inv.WhatsappNumber),
		"type":              "template",
		"template": map[string]interface{}{
			"name":     s.cfg.MetaTemplateNameInvitation,
			"language": map[string]interface{}{"code": s.cfg.MetaTemplateLanguage},
			"components": []map[string]interface{}{
				{
					"type": "header",
					"parameters": []map[string]interface{}{
						{"type": "image", "image": map[string]interface{}{"link": headerImageURL}},
					},
				},
				{
					"type": "body",
					"parameters": []map[string]interface{}{
						{"type": "text", "parameter_name": "guest_name", "text": inv.GuestName},
					},
				},
			},
		},
	}
	return s.send(payload)
}

// send POSTs a message payload to the Meta Cloud API messages endpoint and
// returns a normalized SendResult.
// https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages
func (s *metaWhatsappService) send(payload map[string]interface{}) SendResult {
	if s.cfg.MetaAccessToken == "" || s.cfg.MetaPhoneNumberID == "" {
		return SendResult{Err: fmt.Errorf("meta whatsapp is not configured")}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SendResult{Err: err}
	}

	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", s.cfg.MetaAPIVersion, s.cfg.MetaPhoneNumberID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return SendResult{Err: err}
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.MetaAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return SendResult{Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return SendResult{Err: err}
	}

	var parsed map[string]interface{}
	_ = json.Unmarshal(respBody, &parsed)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SendResult{Response: parsed, Err: fmt.Errorf("meta error (%d): %s", resp.StatusCode, metaErrorMessage(parsed, respBody))}
	}

	return SendResult{
		WhatsappMessageID: metaMessageID(parsed),
		Response:          parsed,
	}
}

// metaMessageID extracts messages[0].id from a successful Meta response.
func metaMessageID(resp map[string]interface{}) string {
	messages, ok := resp["messages"].([]interface{})
	if !ok || len(messages) == 0 {
		return ""
	}
	first, ok := messages[0].(map[string]interface{})
	if !ok {
		return ""
	}
	id, _ := first["id"].(string)
	return id
}

// metaErrorMessage extracts error.message from a Meta error response, falling
// back to the raw body when the shape is unexpected.
func metaErrorMessage(resp map[string]interface{}, rawBody []byte) string {
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		if msg, ok := errObj["message"].(string); ok && msg != "" {
			return msg
		}
	}
	return string(rawBody)
}

// normalizeIndonesianPhone converts a stored WhatsApp number into the digits-
// only E.164 format Meta expects (country code + number, no leading "+", no
// leading zero). Example: "08123456789" -> "628123456789".
func normalizeIndonesianPhone(number string) string {
	n := strings.TrimSpace(number)
	n = strings.TrimPrefix(n, "whatsapp:")
	n = strings.TrimPrefix(n, "+")
	n = strings.ReplaceAll(n, "-", "")
	n = strings.ReplaceAll(n, " ", "")
	if strings.HasPrefix(n, "0") {
		n = "62" + n[1:]
	}
	return n
}

func coupleName(inv *models.Invitation) string {
	if inv.Event != nil && inv.Event.CoupleName != "" {
		return inv.Event.CoupleName
	}
	return "the couple"
}
