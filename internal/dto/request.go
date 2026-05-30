package dto

import "time"

// CreateEventRequest is the body for POST /api/events.
type CreateEventRequest struct {
	Tag                   string     `json:"tag" binding:"required"`
	CoupleName            string     `json:"couple_name" binding:"required"`
	HolyMatrimonyDate     *time.Time `json:"holy_matrimony_date"`
	HolyMatrimonyLocation *string    `json:"holy_matrimony_location"`
	ReceptionDate         *time.Time `json:"reception_date"`
	ReceptionLocation     *string    `json:"reception_location"`
	GiftAddress           *string    `json:"gift_address"`
	BankAccount           *string    `json:"bank_account"`
}

// InvitationFilter holds the query params for GET /api/invitations.
type InvitationFilter struct {
	Tag              string `form:"tag"`
	EventID          uint64 `form:"event_id"`
	RSVPStatus       string `form:"rsvp_status"`
	InvitationStatus string `form:"invitation_status"`
	QRStatus         string `form:"qr_status"`
}

// SendInvitationsRequest is the body for POST /api/invitations/send.
type SendInvitationsRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// TagRequest is the common body for tag-scoped batch operations.
type TagRequest struct {
	Tag string `json:"tag" binding:"required"`
}

// ResendQRRequest is the body for POST /api/invitations/resend-qr.
type ResendQRRequest struct {
	Tag string   `json:"tag" binding:"required"`
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// CheckinRequest is the body for POST /api/check-in.
type CheckinRequest struct {
	QRCodeToken string `json:"qr_code_token" binding:"required"`
	EventType   string `json:"event_type" binding:"required"`
	ActualPax   uint16 `json:"actual_pax" binding:"required"`
	ScannerName string `json:"scanner_name"`
}
