package services

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"

	"projectvows/internal/dto"
	"projectvows/internal/models"
	"projectvows/internal/repositories"
)

type InvitationService struct {
	invRepo  *repositories.InvitationRepository
	logRepo  *repositories.WhatsappLogRepository
	whatsapp WhatsappService
	qr       QRService
}

func NewInvitationService(
	invRepo *repositories.InvitationRepository,
	logRepo *repositories.WhatsappLogRepository,
	whatsapp WhatsappService,
	qr QRService,
) *InvitationService {
	return &InvitationService{
		invRepo:  invRepo,
		logRepo:  logRepo,
		whatsapp: whatsapp,
		qr:       qr,
	}
}

func (s *InvitationService) List(f dto.InvitationFilter) ([]models.Invitation, error) {
	return s.invRepo.List(f)
}

func (s *InvitationService) GetByID(id uint64) (*models.Invitation, error) {
	return s.invRepo.FindByID(id)
}

// QRImageByToken renders the QR PNG for the invitation identified by qrToken.
// This backs the public /qr endpoint that Twilio fetches as message media.
func (s *InvitationService) QRImageByToken(qrToken string) ([]byte, error) {
	inv, err := s.invRepo.FindByQRToken(qrToken)
	if err != nil {
		return nil, err
	}
	return s.qr.Generate(s.qr.CheckinURL(inv.QRCodeToken))
}

// SendByIDs sends the initial invitation to the given invitation ids.
func (s *InvitationService) SendByIDs(ids []uint64) (*dto.BatchSummary, error) {
	invs, err := s.invRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	return s.sendInvitationBatch(invs, models.MsgTypeInitialInvitation), nil
}

// SendPending sends the invitation to all imported (not yet sent) guests of a tag.
func (s *InvitationService) SendPending(tag string) (*dto.BatchSummary, error) {
	invs, err := s.invRepo.FindPendingByTag(tag)
	if err != nil {
		return nil, err
	}
	return s.sendInvitationBatch(invs, models.MsgTypeInitialInvitation), nil
}

// ResendUnanswered resends the invitation to guests who have not answered.
func (s *InvitationService) ResendUnanswered(tag string) (*dto.BatchSummary, error) {
	invs, err := s.invRepo.FindUnansweredByTag(tag)
	if err != nil {
		return nil, err
	}
	return s.sendInvitationBatch(invs, models.MsgTypeResendInvitation), nil
}

// SendReminder sends a reminder to all attending guests.
func (s *InvitationService) SendReminder(tag string) (*dto.BatchSummary, error) {
	invs, err := s.invRepo.FindAttendingByTag(tag)
	if err != nil {
		return nil, err
	}
	summary := &dto.BatchSummary{Total: len(invs), Results: []dto.BatchResult{}}
	for i := range invs {
		inv := &invs[i]
		res := s.whatsapp.SendReminder(inv)
		s.logSend(inv, models.MsgTypeReminder, res)
		s.appendResult(summary, inv.ID, res.Err)
	}
	return summary, nil
}

// sendInvitationBatch sends invitations and updates invitation_status accordingly.
func (s *InvitationService) sendInvitationBatch(invs []models.Invitation, msgType string) *dto.BatchSummary {
	summary := &dto.BatchSummary{Total: len(invs), Results: []dto.BatchResult{}}
	for i := range invs {
		inv := &invs[i]
		res := s.whatsapp.SendInvitation(inv, msgType)
		s.logSend(inv, msgType, res)

		if res.Err != nil {
			inv.InvitationStatus = models.InvitationStatusFailed
		} else {
			inv.InvitationStatus = models.InvitationStatusSent
		}
		_ = s.invRepo.Save(inv)
		s.appendResult(summary, inv.ID, res.Err)
	}
	return summary
}

// GenerateAndSendQR generates QR codes in memory and sends them to eligible
// guests. Nothing is stored: no file, no upload, no QR metadata persisted.
func (s *InvitationService) GenerateAndSendQR(tag string) (*dto.BatchSummary, error) {
	invs, err := s.invRepo.FindQRReadyByTag(tag)
	if err != nil {
		return nil, err
	}
	summary := &dto.BatchSummary{Total: len(invs), Results: []dto.BatchResult{}}
	for i := range invs {
		inv := &invs[i]
		s.processQR(inv, summary, models.MsgTypeSendQR)
	}
	return summary, nil
}

// ResendQR resends QR to selected ids. Every resend generates a fresh QR in
// memory and sends it immediately; nothing is reused or persisted.
func (s *InvitationService) ResendQR(tag string, ids []uint64) (*dto.BatchSummary, error) {
	invs, err := s.invRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	summary := &dto.BatchSummary{Results: []dto.BatchResult{}}
	for i := range invs {
		inv := &invs[i]
		// Guard: only resend within the requested tag.
		if inv.Event == nil || inv.Event.Tag != tag {
			summary.Results = append(summary.Results, dto.BatchResult{
				ID: inv.ID, Success: false, Reason: "invitation does not belong to tag",
			})
			summary.Failed++
			continue
		}
		summary.Total++
		s.processQR(inv, summary, models.MsgTypeResendQR)
	}
	return summary, nil
}

// processQR generates a QR in memory and sends it immediately for one
// invitation. No QR image or QR metadata is stored: the PNG is generated to
// validate encoding (and served on demand by the /qr endpoint that Twilio
// fetches), then discarded once this function returns.
func (s *InvitationService) processQR(inv *models.Invitation, summary *dto.BatchSummary, msgType string) {
	// Generate in memory to validate; the bytes are intentionally discarded.
	if _, err := s.qr.Generate(s.qr.CheckinURL(inv.QRCodeToken)); err != nil {
		s.failQR(inv, summary, "qr generation failed: "+err.Error())
		return
	}

	// Send using the public on-demand URL (Twilio fetches it as media).
	res := s.whatsapp.SendQR(inv, s.qr.ImageURL(inv.QRCodeToken))
	s.logSend(inv, msgType, res)
	s.appendResult(summary, inv.ID, res.Err)
}

func (s *InvitationService) failQR(inv *models.Invitation, summary *dto.BatchSummary, reason string) {
	summary.Failed++
	summary.Results = append(summary.Results, dto.BatchResult{ID: inv.ID, Success: false, Reason: reason})
}

// logSend writes a whatsapp_logs row for an outbound send result.
func (s *InvitationService) logSend(inv *models.Invitation, msgType string, res SendResult) {
	log := &models.WhatsappLog{
		InvitationID: inv.ID,
		MessageType:  msgType,
		Direction:    models.DirectionOutbound,
	}
	if res.WhatsappMessageID != "" {
		log.WhatsappMessageID = &res.WhatsappMessageID
	}
	if res.Err != nil {
		log.Status = models.WALogFailed
		log.MetaError = toJSON(map[string]string{"error": res.Err.Error()})
	} else {
		now := time.Now()
		log.Status = models.WALogSent
		log.SentAt = &now
		log.MetaResponse = toJSON(res.Response)
	}
	_ = s.logRepo.Create(log)
}

func (s *InvitationService) appendResult(summary *dto.BatchSummary, id uint64, err error) {
	if err != nil {
		summary.Failed++
		summary.Results = append(summary.Results, dto.BatchResult{ID: id, Success: false, Reason: err.Error()})
		return
	}
	summary.Succeeded++
	summary.Results = append(summary.Results, dto.BatchResult{ID: id, Success: true})
}

func toJSON(v interface{}) datatypes.JSON {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return datatypes.JSON(b)
}
