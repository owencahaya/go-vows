package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"projectvows/internal/config"
	"projectvows/internal/handlers"
	"projectvows/internal/repositories"
	"projectvows/internal/routes"
	"projectvows/internal/services"
)

func main() {
	// Load .env if present (ignored in production where env vars are injected).
	_ = godotenv.Load()

	cfg := config.Load()

	// Database
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := config.Migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Repositories
	eventRepo := repositories.NewEventRepository(db)
	invRepo := repositories.NewInvitationRepository(db)
	logRepo := repositories.NewWhatsappLogRepository(db)
	checkinRepo := repositories.NewCheckinLogRepository(db)

	// External services
	whatsappSvc := services.NewWhatsappService(cfg)
	qrSvc := services.NewQRService(cfg)

	// Domain services
	eventSvc := services.NewEventService(eventRepo)
	csvSvc := services.NewCSVService(db, eventRepo)
	invSvc := services.NewInvitationService(invRepo, logRepo, whatsappSvc, qrSvc)
	checkinSvc := services.NewCheckinService(invRepo, checkinRepo)
	webhookSvc := services.NewWebhookService(cfg, invRepo, logRepo, whatsappSvc, invSvc)

	// Handlers
	h := routes.Handlers{
		Event:      handlers.NewEventHandler(eventSvc),
		Invitation: handlers.NewInvitationHandler(invSvc, csvSvc),
		Webhook:    handlers.NewWebhookHandler(webhookSvc),
		Checkin:    handlers.NewCheckinHandler(checkinSvc),
	}

	// Router
	r := gin.Default()
	routes.Register(r, h)

	addr := ":" + cfg.AppPort
	log.Printf("Project Vows API listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
