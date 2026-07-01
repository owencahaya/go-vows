package routes

import (
	"github.com/gin-gonic/gin"

	"projectvows/internal/handlers"
)

// Handlers bundles the HTTP handlers wired into the router.
type Handlers struct {
	Event      *handlers.EventHandler
	Invitation *handlers.InvitationHandler
	Webhook    *handlers.WebhookHandler
	Checkin    *handlers.CheckinHandler
}

// Register mounts all API routes onto the given engine.
func Register(r *gin.Engine, h Handlers) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public QR image, generated in memory on demand (nothing is stored) and
	// fetched by Twilio as WhatsApp media. URL matches QRService.ImageURL.
	r.GET("/qr/:qr_code_token", h.Invitation.QRImage)

	api := r.Group("/api")

	// Events
	api.POST("/events", h.Event.Create)
	api.GET("/events", h.Event.List)
	api.GET("/events/:id", h.Event.Get)

	// Invitations
	api.POST("/invitations/import-csv", h.Invitation.ImportCSV)
	api.GET("/invitations", h.Invitation.List)
	api.GET("/invitations/:id", h.Invitation.Get)

	// WhatsApp send operations
	api.POST("/invitations/send", h.Invitation.Send)
	api.POST("/invitations/send-pending", h.Invitation.SendPending)
	api.POST("/invitations/resend-unanswered", h.Invitation.ResendUnanswered)
	api.POST("/invitations/send-reminder", h.Invitation.SendReminder)
	api.POST("/invitations/generate-send-qr", h.Invitation.GenerateSendQR)
	api.POST("/invitations/resend-qr", h.Invitation.ResendQR)

	// WhatsApp webhook
	api.GET("/webhook/whatsapp", h.Webhook.Verify)
	api.POST("/webhook/whatsapp", h.Webhook.Receive)

	// Check-in
	api.GET("/check-in/:qr_code_token", h.Checkin.Lookup)
	api.POST("/check-in", h.Checkin.Checkin)
}
