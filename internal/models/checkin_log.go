package models

import "time"

// Check-in event types.
const (
	CheckinEventHolyMatrimony = "holy_matrimony"
	CheckinEventReception     = "reception"
)

// CheckinLog stores QR scan/check-in logs per event type.
type CheckinLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	InvitationID uint64    `gorm:"not null;uniqueIndex:uniq_invitation_event" json:"invitation_id"`
	EventType    string    `gorm:"type:varchar(30);not null;uniqueIndex:uniq_invitation_event" json:"event_type"`
	CheckedInPax uint16    `gorm:"type:smallint unsigned;not null" json:"checked_in_pax"`
	CheckedInAt  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"checked_in_at"`
	ScannerName  *string   `gorm:"type:varchar(150)" json:"scanner_name"`
	CreatedAt    time.Time `json:"created_at"`
}

func (CheckinLog) TableName() string { return "checkin_logs" }
