package services

import (
	"bytes"
	"image/png"
	"strings"

	"github.com/skip2/go-qrcode"

	"projectvows/internal/config"
)

// QRService generates QR code images for invitations.
type QRService interface {
	// Generate returns PNG bytes encoding the given content (e.g. a check-in URL).
	Generate(content string) ([]byte, error)
	// CheckinURL builds the URL that the QR encodes for a given token.
	CheckinURL(qrToken string) string
	// ImageURL builds the public URL where the QR PNG is served on demand.
	// Meta fetches this as image.link message media; nothing is stored server-side.
	ImageURL(qrToken string) string
}

type qrService struct {
	cfg *config.Config
}

// NewQRService returns a QR generator. This uses a real, lightweight PNG
// encoder so the rest of the pipeline has actual bytes to upload; swap the
// content/branding later as needed.
func NewQRService(cfg *config.Config) QRService {
	return &qrService{cfg: cfg}
}

func (s *qrService) CheckinURL(qrToken string) string {
	return s.cfg.AppBaseURL + "/check-in/" + qrToken
}

func (s *qrService) ImageURL(qrToken string) string {
	return strings.TrimRight(s.cfg.AppBaseURL, "/") + "/qr/" + qrToken + ".png"
}

func (s *qrService) Generate(content string) ([]byte, error) {
	// TODO: add couple branding / logo overlay if desired.
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	img := qr.Image(512)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
