package services

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"

	"gorm.io/gorm"

	"projectvows/internal/dto"
	"projectvows/internal/models"
	"projectvows/internal/repositories"
	"projectvows/internal/utils"
)

// CSVService handles parsing and importing guest CSV files.
type CSVService struct {
	db        *gorm.DB
	eventRepo *repositories.EventRepository
}

func NewCSVService(db *gorm.DB, eventRepo *repositories.EventRepository) *CSVService {
	return &CSVService{db: db, eventRepo: eventRepo}
}

// Import parses the CSV reader and inserts invitations.
//
// Expected header: tag,guest_name,whatsapp_number
// Each row is validated and imported independently; failures are collected and
// returned in the summary instead of aborting the whole import.
func (s *CSVService) Import(r io.Reader) (*dto.ImportSummary, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // tolerate ragged rows; we validate manually

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("empty CSV file")
	}

	// Detect and skip a header row if present.
	start := 0
	if isHeader(records[0]) {
		start = 1
	}

	summary := &dto.ImportSummary{FailedRows: []dto.ImportFailedRow{}}
	// Cache events by tag to avoid repeated lookups.
	eventCache := map[string]*models.Event{}

	for i := start; i < len(records); i++ {
		rowNum := i + 1 // 1-based for human readability
		row := records[i]
		summary.TotalRows++

		tag, guestName, waNumber, verr := parseRow(row)
		if verr != "" {
			summary.FailedCount++
			summary.FailedRows = append(summary.FailedRows, dto.ImportFailedRow{Row: rowNum, Reason: verr})
			continue
		}

		event, ok := eventCache[tag]
		if !ok {
			ev, err := s.eventRepo.FindByTag(tag)
			if err != nil {
				summary.FailedCount++
				summary.FailedRows = append(summary.FailedRows, dto.ImportFailedRow{
					Row: rowNum, Reason: "event not found for tag", Tag: tag, Guest: guestName,
				})
				continue
			}
			event = ev
			eventCache[tag] = ev
		}

		inv := &models.Invitation{
			EventID:        event.ID,
			GuestName:      guestName,
			WhatsappNumber: waNumber,
			InvitationCode: utils.NewInvitationCode(),
			QRCodeToken:    utils.NewQRToken(),
		}

		if err := s.db.Create(inv).Error; err != nil {
			reason := "failed to insert"
			if isDuplicateErr(err) {
				reason = "duplicate guest (event_id + whatsapp_number already exists)"
			}
			summary.FailedCount++
			summary.FailedRows = append(summary.FailedRows, dto.ImportFailedRow{
				Row: rowNum, Reason: reason, Tag: tag, Guest: guestName,
			})
			continue
		}
		summary.SuccessCount++
	}

	return summary, nil
}

func parseRow(row []string) (tag, guestName, waNumber, errReason string) {
	if len(row) < 3 {
		return "", "", "", "expected 3 columns: tag,guest_name,whatsapp_number"
	}
	tag = strings.TrimSpace(row[0])
	guestName = strings.TrimSpace(row[1])
	waNumber = strings.TrimSpace(row[2])
	if tag == "" {
		return "", "", "", "missing tag"
	}
	if guestName == "" {
		return "", "", "", "missing guest_name"
	}
	if waNumber == "" {
		return "", "", "", "missing whatsapp_number"
	}
	return tag, guestName, waNumber, ""
}

func isHeader(row []string) bool {
	if len(row) < 1 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(row[0]), "tag")
}

func isDuplicateErr(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	// Fallback: MySQL duplicate-entry text (driver may not map to ErrDuplicatedKey).
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
