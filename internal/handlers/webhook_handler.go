package handlers

import (
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
// Twilio does not use a GET verification handshake (unlike Meta). This is kept
// for backward compatibility / health probing and simply returns 200 OK.
func (h *WebhookHandler) Verify(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// POST /api/webhook/whatsapp
// Receives Twilio delivery/status callbacks and inbound guest messages
// (application/x-www-form-urlencoded), authenticated via X-Twilio-Signature.
func (h *WebhookHandler) Receive(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		utils.BadRequest(c, "cannot parse form body")
		return
	}

	signature := c.GetHeader("X-Twilio-Signature")
	if !h.svc.ValidateSignature(fullRequestURL(c), c.Request.PostForm, signature) {
		utils.Error(c, http.StatusForbidden, "invalid_signature", "invalid Twilio signature")
		return
	}

	if err := h.svc.HandleTwilioCallback(c.Request.PostForm); err != nil {
		utils.InternalError(c, err)
		return
	}

	// Twilio expects a fast 2xx acknowledgement.
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// fullRequestURL reconstructs the absolute URL Twilio signed the request with,
// honoring reverse-proxy forwarding headers when present.
func fullRequestURL(c *gin.Context) string {
	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	if fwdHost := c.GetHeader("X-Forwarded-Host"); fwdHost != "" {
		host = fwdHost
	}
	return scheme + "://" + host + c.Request.URL.RequestURI()
}
