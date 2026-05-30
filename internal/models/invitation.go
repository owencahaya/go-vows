package models

import "time"

// Invitation status (delivery of the initial invitation message).
const (
	InvitationStatusImported = "imported"
	InvitationStatusSent     = "invitation_sent"
	InvitationStatusFailed   = "invitation_failed"
)

// RSVP status.
const (
	RSVPNotAnswered  = "not_answered"
	RSVPAttending    = "attending"
	RSVPNotAttending = "not_attending"
)

// Event choice.
const (
	EventChoiceHolyMatrimony = "holy_matrimony"
	EventChoiceReception     = "reception"
	EventChoiceBoth          = "both"
)

// Gift interest.
const (
	GiftInterestNotAsked = "not_asked"
	GiftInterestYes      = "yes"
	GiftInterestNo       = "no"
)

// QR status.
const (
	QRStatusNotGenerated = "not_generated"
	QRStatusGenerated    = "generated"
	QRStatusSent         = "sent"
	QRStatusSendFailed   = "send_failed"
)

// Invitation stores guest invitation, RSVP, pax, event choice and QR metadata.
type Invitation struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID        uint64 `gorm:"not null;index;uniqueIndex:uniq_event_whatsapp" json:"event_id"`
	GuestName      string `gorm:"type:varchar(150);not null" json:"guest_name"`
	WhatsappNumber string `gorm:"type:varchar(30);not null;uniqueIndex:uniq_event_whatsapp" json:"whatsapp_number"`
	InvitationCode string `gorm:"type:varchar(100);uniqueIndex;not null" json:"invitation_code"`

	InvitationStatus string  `gorm:"type:varchar(30);not null;default:imported;index" json:"invitation_status"`
	RSVPStatus       string  `gorm:"column:rsvp_status;type:varchar(30);not null;default:not_answered;index" json:"rsvp_status"`
	PaxCount         *uint16 `gorm:"type:smallint unsigned" json:"pax_count"`
	EventChoice      *string `gorm:"type:varchar(30);index" json:"event_choice"`
	GiftInterest     string  `gorm:"type:varchar(30);not null;default:not_asked" json:"gift_interest"`

	QRCodeToken       string     `gorm:"column:qr_code_token;type:varchar(150);uniqueIndex;not null" json:"qr_code_token"`
	QRWhatsappMediaID *string    `gorm:"column:qr_whatsapp_media_id;type:varchar(255)" json:"qr_whatsapp_media_id"`
	QRMediaUploadedAt *time.Time `gorm:"column:qr_media_uploaded_at;type:datetime" json:"qr_media_uploaded_at"`
	QRSentAt          *time.Time `gorm:"column:qr_sent_at;type:datetime" json:"qr_sent_at"`
	QRStatus          string     `gorm:"column:qr_status;type:varchar(30);not null;default:not_generated;index" json:"qr_status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Event        *Event        `gorm:"foreignKey:EventID" json:"event,omitempty"`
	CheckinLogs  []CheckinLog  `gorm:"foreignKey:InvitationID" json:"checkin_logs,omitempty"`
	WhatsappLogs []WhatsappLog `gorm:"foreignKey:InvitationID" json:"-"`
}

func (Invitation) TableName() string { return "invitations" }
