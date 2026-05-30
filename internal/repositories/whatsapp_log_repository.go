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

// WithTx returns a repository bound to the given transaction.
func (r *WhatsappLogRepository) WithTx(tx *gorm.DB) *WhatsappLogRepository {
	return &WhatsappLogRepository{db: tx}
}
