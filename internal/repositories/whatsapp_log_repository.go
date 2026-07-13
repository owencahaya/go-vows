package repositories

import (
	"gorm.io/gorm"

	"projectvows/internal/models"
)

type WhatsappLogRepository struct {
	db *gorm.DB
}

func NewWhatsappLogRepository(db *gorm.DB) *WhatsappLogRepository {
	return &WhatsappLogRepository{db: db}
}

func (r *WhatsappLogRepository) Create(log *models.WhatsappLog) error {
	return r.db.Create(log).Error
}

// UpdateStatusByMessageID updates the status of the most recent outbound log
// matching the given provider message id (Meta wamid). Used by the Meta
// webhook status update. It is a no-op (nil error) when no log matches.
func (r *WhatsappLogRepository) UpdateStatusByMessageID(messageID, status string) error {
	return r.db.Model(&models.WhatsappLog{}).
		Where("whatsapp_message_id = ?", messageID).
		Update("status", status).Error
}

// WithTx returns a repository bound to the given transaction.
func (r *WhatsappLogRepository) WithTx(tx *gorm.DB) *WhatsappLogRepository {
	return &WhatsappLogRepository{db: tx}
}
