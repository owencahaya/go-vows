package services

import (
	"testing"

	"projectvows/internal/config"
	"projectvows/internal/models"
)

func TestToWhatsAppAddress(t *testing.T) {
	cases := map[string]string{
		"+14155238886":          "whatsapp:+14155238886",
		"14155238886":           "whatsapp:+14155238886",
		"whatsapp:+14155238886": "whatsapp:+14155238886",
		" +6281234567 ":         "whatsapp:+6281234567",
		"":                      "",
	}
	for in, want := range cases {
		if got := toWhatsAppAddress(in); got != want {
			t.Errorf("toWhatsAppAddress(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapTwilioStatus(t *testing.T) {
	cases := map[string]string{
		"failed":      models.WALogFailed,
		"undelivered": models.WALogFailed,
		"sent":        models.WALogSent,
		"delivered":   models.WALogSent,
		"read":        models.WALogSent,
		"queued":      models.WALogPending,
		"sending":     models.WALogPending,
	}
	for in, want := range cases {
		if got := mapTwilioStatus(in); got != want {
			t.Errorf("mapTwilioStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

// When Twilio is unconfigured, sends must fail fast without hitting the network.
func TestSendRequiresConfig(t *testing.T) {
	svc := NewWhatsappService(&config.Config{})
	inv := &models.Invitation{WhatsappNumber: "+14155238886"}

	if res := svc.SendInvitation(inv, models.MsgTypeInitialInvitation); res.Err == nil {
		t.Error("SendInvitation with empty config should return an error")
	}
	if res := svc.SendQR(inv, "https://example.com/qr/abc.png"); res.Err == nil {
		t.Error("SendQR with empty config should return an error")
	}
}
