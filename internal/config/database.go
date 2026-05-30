package config

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"projectvows/internal/models"
)

// NewDatabase opens a GORM connection to MySQL and configures the pool.
func NewDatabase(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		// Map driver errors to gorm.ErrDuplicatedKey / ErrRecordNotFound etc.
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Migrate runs GORM AutoMigrate for all models.
func Migrate(db *gorm.DB) error {
	log.Println("running database migrations...")
	return db.AutoMigrate(
		&models.Event{},
		&models.Invitation{},
		&models.WhatsappLog{},
		&models.CheckinLog{},
	)
}
