package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"projectvows/internal/dto"
	"projectvows/internal/services"
	"projectvows/internal/utils"
)

type CheckinHandler struct {
	svc *services.CheckinService
}

func NewCheckinHandler(svc *services.CheckinService) *CheckinHandler {
	return &CheckinHandler{svc: svc}
}

// GET /api/check-in/:qr_code_token
func (h *CheckinHandler) Lookup(c *gin.Context) {
	token := c.Param("qr_code_token")
	resp, err := h.svc.Lookup(token)
	if err != nil {
		if errors.Is(err, services.ErrInvitationNotFound) {
			utils.Error(c, http.StatusNotFound, "invitation_not_found", "invitation not found")
			return
		}
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "", resp)
}

// POST /api/check-in
func (h *CheckinHandler) Checkin(c *gin.Context) {
	var req dto.CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	inv, log, err := h.svc.Checkin(req)
	if err != nil {
		h.writeCheckinError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, "Check-in berhasil", gin.H{
		"name":           inv.GuestName,
		"event_type":     log.EventType,
		"registered_pax": inv.PaxCount,
		"checked_in_pax": log.CheckedInPax,
	})
}

// writeCheckinError maps domain errors to clear HTTP responses.
func (h *CheckinHandler) writeCheckinError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvitationNotFound):
		utils.Error(c, http.StatusNotFound, "invitation_not_found", "invitation not found")
	case errors.Is(err, services.ErrNotAttending):
		utils.Error(c, http.StatusUnprocessableEntity, "not_attending", "guest is not attending")
	case errors.Is(err, services.ErrInvalidEvent):
		utils.Error(c, http.StatusUnprocessableEntity, "invalid_event", "event_choice does not allow this event_type")
	case errors.Is(err, services.ErrInvalidPax):
		utils.Error(c, http.StatusUnprocessableEntity, "invalid_pax", "actual_pax exceeds registered pax_count")
	case errors.Is(err, services.ErrAlreadyCheckedIn):
		utils.Error(c, http.StatusConflict, "already_checked_in", "guest already checked in for this event")
	default:
		utils.InternalError(c, err)
	}
}
