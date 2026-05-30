package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"projectvows/internal/services"
	"projectvows/internal/utils"
)

type WebhookHandler struct {
	svc *services.WebhookService
}

func NewWebhookHandler(svc *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

// GET /api/webhook/whatsapp
// Meta verification handshake: echoes hub.challenge when the token matches.
func (h *WebhookHandler) Verify(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if resp, ok := h.svc.VerifyToken(mode, token, challenge); ok {
		c.String(http.StatusOK, resp)
		return
	}
	utils.Error(c, http.StatusForbidden, "verification_failed", "invalid verify token")
}

// POST /api/webhook/whatsapp
// Receives inbound WhatsApp events.
func (h *WebhookHandler) Receive(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.BadRequest(c, "cannot read body")
		return
	}

	// TODO: parse Meta interactive button/list replies and route to the
	// WebhookService conversation mutators (UpdateRSVPStatus, UpdatePaxCount,
	// UpdateEventChoice, UpdateGiftInterest, TriggerNextMessage, TriggerQRGeneration).
	if err := h.svc.HandleInbound(body); err != nil {
		utils.InternalError(c, err)
		return
	}

	// Meta expects a fast 200 OK acknowledgement.
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}
