package models

import (
	"time"

	"gorm.io/datatypes"
)

// WhatsApp message types.
const (
	MsgTypeInitialInvitation = "initial_invitation"
	MsgTypeResendInvitation  = "resend_invitation"
	MsgTypeAskPax            = "ask_pax"
	MsgTypeAskEventChoice    = "ask_event_choice"
	MsgTypeSendQR            = "send_qr"
	MsgTypeResendQR          = "resend_qr"
	MsgTypeReminder          = "reminder"
	MsgTypeGiftInfo          = "gift_info"
	MsgTypeOther             = "other"
)

// Message direction.
const (
	DirectionOutbound = "outbound"
	DirectionInbound  = "inbound"
)

// WhatsApp log status.
const (
	WALogPending  = "pending"
	WALogSent     = "sent"
	WALogFailed   = "failed"
	WALogReceived = "received"
)

// WhatsappLog stores WhatsApp inbound/outbound logs and Meta API responses/errors.
type WhatsappLog struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	InvitationID      uint64         `gorm:"not null;index" json:"invitation_id"`
	MessageType       string         `gorm:"type:varchar(40);index" json:"message_type"`
	Direction         string         `gorm:"type:varchar(20)" json:"direction"`
	WhatsappMessageID *string        `gorm:"column:whatsapp_message_id;type:varchar(255);index" json:"whatsapp_message_id"`
	Status            string         `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
	Payload           datatypes.JSON `gorm:"type:json" json:"payload"`
	MetaResponse      datatypes.JSON `gorm:"type:json" json:"meta_response"`
	MetaError         datatypes.JSON `gorm:"type:json" json:"meta_error"`
	SentAt            *time.Time     `gorm:"type:datetime" json:"sent_at"`
	ReceivedAt        *time.Time     `gorm:"type:datetime" json:"received_at"`
	CreatedAt         time.Time      `json:"created_at"`
}

func (WhatsappLog) TableName() string { return "whatsapp_logs" }
