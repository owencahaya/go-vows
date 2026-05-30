package utils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

// NewInvitationCode returns a short, URL-safe random invitation code.
func NewInvitationCode() string {
	return "inv_" + randomHex(8)
}

// NewQRToken returns a UUID v4 used as the QR check-in token.
func NewQRToken() string {
	return uuid.NewString()
}

// randomHex returns n random bytes encoded as a hex string.
// Falls back to a UUID if the crypto source fails.
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return uuid.NewString()
	}
	return hex.EncodeToString(b)
}
