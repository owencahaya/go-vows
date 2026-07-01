package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"projectvows/internal/dto"
	"projectvows/internal/repositories"
	"projectvows/internal/services"
	"projectvows/internal/utils"
)

type InvitationHandler struct {
	svc    *services.InvitationService
	csvSvc *services.CSVService
}

func NewInvitationHandler(svc *services.InvitationService, csvSvc *services.CSVService) *InvitationHandler {
	return &InvitationHandler{svc: svc, csvSvc: csvSvc}
}

// POST /api/invitations/import-csv
func (h *InvitationHandler) ImportCSV(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "missing form field 'file'")
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	defer f.Close()

	summary, err := h.csvSvc.Import(f)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, http.StatusOK, "Import completed", summary)
}

// GET /api/invitations
func (h *InvitationHandler) List(c *gin.Context) {
	var filter dto.InvitationFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	invs, err := h.svc.List(filter)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "", invs)
}

// GET /api/invitations/:id
func (h *InvitationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid invitation id")
		return
	}
	inv, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, "invitation_not_found", "invitation not found")
			return
		}
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "", inv)
}

// POST /api/invitations/send
func (h *InvitationHandler) Send(c *gin.Context) {
	var req dto.SendInvitationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	summary, err := h.svc.SendByIDs(req.IDs)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "Invitations processed", summary)
}

// POST /api/invitations/send-pending
func (h *InvitationHandler) SendPending(c *gin.Context) {
	h.runTagBatch(c, func(tag string) (*dto.BatchSummary, error) {
		return h.svc.SendPending(tag)
	})
}

// POST /api/invitations/resend-unanswered
func (h *InvitationHandler) ResendUnanswered(c *gin.Context) {
	h.runTagBatch(c, func(tag string) (*dto.BatchSummary, error) {
		return h.svc.ResendUnanswered(tag)
	})
}

// POST /api/invitations/send-reminder
func (h *InvitationHandler) SendReminder(c *gin.Context) {
	h.runTagBatch(c, func(tag string) (*dto.BatchSummary, error) {
		return h.svc.SendReminder(tag)
	})
}

// POST /api/invitations/generate-send-qr
func (h *InvitationHandler) GenerateSendQR(c *gin.Context) {
	h.runTagBatch(c, func(tag string) (*dto.BatchSummary, error) {
		return h.svc.GenerateAndSendQR(tag)
	})
}

// POST /api/invitations/resend-qr
func (h *InvitationHandler) ResendQR(c *gin.Context) {
	var req dto.ResendQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	summary, err := h.svc.ResendQR(req.Tag, req.IDs)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "QR resend processed", summary)
}

// GET /qr/:qr_code_token
// Serves the QR PNG publicly so Twilio can fetch it as message media.
// Accepts an optional ".png" suffix on the token for a friendly URL/extension.
func (h *InvitationHandler) QRImage(c *gin.Context) {
	token := strings.TrimSuffix(c.Param("qr_code_token"), ".png")
	png, err := h.svc.QRImageByToken(token)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, "invitation_not_found", "invitation not found")
			return
		}
		utils.InternalError(c, err)
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}

// runTagBatch parses a {tag} body and runs the given batch function.
func (h *InvitationHandler) runTagBatch(c *gin.Context, fn func(tag string) (*dto.BatchSummary, error)) {
	var req dto.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	summary, err := fn(req.Tag)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "Batch processed", summary)
}
