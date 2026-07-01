package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"

	"projectvows/internal/config"
	"projectvows/internal/models"
)

// SendResult is the normalized outcome of a WhatsApp send operation.
// Field names are preserved from the previous integration so callers and the
// whatsapp_logs writer stay unchanged; the values now come from Twilio.
type SendResult struct {
	WhatsappMessageID string                 // Twilio Message SID (SM.../MM...)
	Response          map[string]interface{} // raw Twilio response (for logging)
	Err               error                  // non-nil on failure
}

// WhatsappService abstracts the WhatsApp provider (now Twilio). The interface
// is unchanged from the previous Meta-based abstraction.
type WhatsappService interface {
	// SendInvitation sends the initial/resend invitation message.
	SendInvitation(inv *models.Invitation, messageType string) SendResult
	// SendReminder sends a reminder to an attending guest.
	SendReminder(inv *models.Invitation) SendResult
	// SendQR sends the QR image. mediaURL is a publicly reachable HTTPS URL
	// that Twilio fetches (replacing Meta's upload-and-media-id model).
	SendQR(inv *models.Invitation, mediaURL string) SendResult
	// SendText sends a plain text message (used by the webhook conversation flow).
	SendText(inv *models.Invitation, body string) SendResult
}

type twilioWhatsappService struct {
	cfg    *config.Config
	client *twilio.RestClient
}

// NewWhatsappService returns the Twilio-backed implementation. The constructor
// signature is unchanged so the DI wiring in main.go stays the same.
func NewWhatsappService(cfg *config.Config) WhatsappService {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.TwilioAccountSID,
		Password: cfg.TwilioAuthToken,
	})
	return &twilioWhatsappService{cfg: cfg, client: client}
}

func (s *twilioWhatsappService) SendInvitation(inv *models.Invitation, messageType string) SendResult {
	body := fmt.Sprintf(
		"Halo %s, Anda diundang ke acara pernikahan %s. Mohon konfirmasi kehadiran Anda.",
		inv.GuestName, coupleName(inv),
	)
	return s.send(inv, sendOptions{
		body:        body,
		contentSid:  s.cfg.TwilioContentSidInvitation,
		contentVars: map[string]string{"1": inv.GuestName, "2": coupleName(inv)},
	})
}

func (s *twilioWhatsappService) SendReminder(inv *models.Invitation) SendResult {
	body := fmt.Sprintf(
		"Halo %s, ini pengingat untuk acara pernikahan %s. Sampai jumpa!",
		inv.GuestName, coupleName(inv),
	)
	return s.send(inv, sendOptions{
		body:        body,
		contentSid:  s.cfg.TwilioContentSidReminder,
		contentVars: map[string]string{"1": inv.GuestName, "2": coupleName(inv)},
	})
}

func (s *twilioWhatsappService) SendQR(inv *models.Invitation, mediaURL string) SendResult {
	body := fmt.Sprintf(
		"Halo %s, berikut QR code untuk check-in di acara %s.",
		inv.GuestName, coupleName(inv),
	)
	return s.send(inv, sendOptions{
		body:        body,
		mediaURL:    mediaURL,
		contentSid:  s.cfg.TwilioContentSidQR,
		contentVars: map[string]string{"1": inv.GuestName},
	})
}

func (s *twilioWhatsappService) SendText(inv *models.Invitation, body string) SendResult {
	return s.send(inv, sendOptions{body: body})
}

// sendOptions describes one outbound message.
type sendOptions struct {
	body        string            // freeform text (used when contentSid is empty)
	mediaURL    string            // optional public media URL (freeform sends only)
	contentSid  string            // optional approved Content Template SID
	contentVars map[string]string // variables for the Content Template
}

// send builds and dispatches the Twilio message, returning a normalized result.
func (s *twilioWhatsappService) send(inv *models.Invitation, opts sendOptions) SendResult {
	if s.cfg.TwilioAccountSID == "" || s.cfg.TwilioAuthToken == "" || s.cfg.TwilioWhatsAppNumber == "" {
		return SendResult{Err: fmt.Errorf("twilio is not configured")}
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetFrom(toWhatsAppAddress(s.cfg.TwilioWhatsAppNumber))
	params.SetTo(toWhatsAppAddress(inv.WhatsappNumber))
	if s.cfg.TwilioStatusCallbackURL != "" {
		params.SetStatusCallback(s.cfg.TwilioStatusCallbackURL)
	}

	if opts.contentSid != "" {
		// Approved template path (production business-initiated messages).
		params.SetContentSid(opts.contentSid)
		if len(opts.contentVars) > 0 {
			if vars, err := json.Marshal(opts.contentVars); err == nil {
				params.SetContentVariables(string(vars))
			}
		}
	} else {
		// Freeform path (Twilio sandbox / open 24h session window).
		params.SetBody(opts.body)
		if opts.mediaURL != "" {
			params.SetMediaUrl([]string{opts.mediaURL})
		}
	}

	resp, err := s.client.Api.CreateMessage(params)
	if err != nil {
		return SendResult{Err: err}
	}

	res := SendResult{Response: twilioResponseMap(resp)}
	if resp.Sid != nil {
		res.WhatsappMessageID = *resp.Sid
	}
	// Twilio may accept the request but report a message-level error code.
	if resp.ErrorCode != nil && *resp.ErrorCode != 0 {
		msg := ""
		if resp.ErrorMessage != nil {
			msg = *resp.ErrorMessage
		}
		res.Err = fmt.Errorf("twilio error %d: %s", *resp.ErrorCode, msg)
	}
	return res
}

// twilioResponseMap flattens the useful Twilio response fields for logging.
func twilioResponseMap(m *twilioApi.ApiV2010Message) map[string]interface{} {
	out := map[string]interface{}{"provider": "twilio"}
	if m == nil {
		return out
	}
	if m.Sid != nil {
		out["sid"] = *m.Sid
	}
	if m.Status != nil {
		out["status"] = *m.Status
	}
	if m.ErrorCode != nil {
		out["error_code"] = *m.ErrorCode
	}
	if m.ErrorMessage != nil {
		out["error_message"] = *m.ErrorMessage
	}
	if m.DateSent != nil {
		out["date_sent"] = *m.DateSent
	}
	return out
}

// toWhatsAppAddress normalizes a phone number into Twilio's "whatsapp:+E164" form.
func toWhatsAppAddress(number string) string {
	n := strings.TrimSpace(number)
	if n == "" {
		return n
	}
	if strings.HasPrefix(n, "whatsapp:") {
		return n
	}
	if !strings.HasPrefix(n, "+") {
		n = "+" + n
	}
	return "whatsapp:" + n
}

func coupleName(inv *models.Invitation) string {
	if inv.Event != nil && inv.Event.CoupleName != "" {
		return inv.Event.CoupleName
	}
	return "the couple"
}
