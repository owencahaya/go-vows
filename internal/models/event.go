package models

import "time"

// Event stores wedding/couple data.
type Event struct {
	ID                    uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Tag                   string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"tag"`
	CoupleName            string     `gorm:"type:varchar(150);not null" json:"couple_name"`
	HolyMatrimonyDate     *time.Time `gorm:"type:datetime" json:"holy_matrimony_date"`
	HolyMatrimonyLocation *string    `gorm:"type:text" json:"holy_matrimony_location"`
	ReceptionDate         *time.Time `gorm:"type:datetime" json:"reception_date"`
	ReceptionLocation     *string    `gorm:"type:text" json:"reception_location"`
	GiftAddress           *string    `gorm:"type:text" json:"gift_address"`
	BankAccount           *string    `gorm:"type:text" json:"bank_account"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`

	Invitations []Invitation `gorm:"foreignKey:EventID" json:"-"`
}

func (Event) TableName() string { return "events" }
