package services

import (
	"testing"

	"projectvows/internal/config"
	"projectvows/internal/models"
)

func TestNormalizeIndonesianPhone(t *testing.T) {
	cases := map[string]string{
		"08123456789":       "628123456789",
		"628123456789":      "628123456789",
		"+628123456789":     "628123456789",
		"whatsapp:+6281234": "6281234",
		" 0812-3456-789 ":   "628123456789",
		"":                  "",
	}
	for in, want := range cases {
		if got := normalizeIndonesianPhone(in); got != want {
			t.Errorf("normalizeIndonesianPhone(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapMetaStatus(t *testing.T) {
	cases := map[string]string{
		"failed":    models.WALogFailed,
		"sent":      models.WALogSent,
		"delivered": models.WALogSent,
		"read":      models.WALogSent,
		"unknown":   models.WALogPending,
	}
	for in, want := range cases {
		if got := mapMetaStatus(in); got != want {
			t.Errorf("mapMetaStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

// When Meta is unconfigured, sends must fail fast without hitting the network.
func TestSendRequiresConfig(t *testing.T) {
	svc := NewWhatsappService(&config.Config{})
	inv := &models.Invitation{WhatsappNumber: "+14155238886"}

	if res := svc.SendInvitation(inv, models.MsgTypeInitialInvitation, ""); res.Err == nil {
		t.Error("SendInvitation with empty config should return an error")
	}
	if res := svc.SendQR(inv, "https://example.com/qr/abc.png"); res.Err == nil {
		t.Error("SendQR with empty config should return an error")
	}
}
