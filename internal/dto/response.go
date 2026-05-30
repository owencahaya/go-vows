package dto

import "time"

// ImportFailedRow describes a CSV row that could not be imported.
type ImportFailedRow struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
	Tag    string `json:"tag,omitempty"`
	Guest  string `json:"guest_name,omitempty"`
}

// ImportSummary is returned by POST /api/invitations/import-csv.
type ImportSummary struct {
	TotalRows    int               `json:"total_rows"`
	SuccessCount int               `json:"success_count"`
	FailedCount  int               `json:"failed_count"`
	FailedRows   []ImportFailedRow `json:"failed_rows"`
}

// BatchResult is the per-id outcome of a batch send/qr operation.
type BatchResult struct {
	ID      uint64 `json:"id"`
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}

// BatchSummary aggregates a batch operation result.
type BatchSummary struct {
	Total     int           `json:"total"`
	Succeeded int           `json:"succeeded"`
	Failed    int           `json:"failed"`
	Results   []BatchResult `json:"results"`
}

// CheckinEventStatus is one event row in the check-in lookup response.
type CheckinEventStatus struct {
	EventType   string     `json:"event_type"`
	CheckedInAt *time.Time `json:"checked_in_at"`
}

// CheckinLookupResponse is returned by GET /api/check-in/:qr_code_token.
type CheckinLookupResponse struct {
	GuestName       string               `json:"guest_name"`
	PaxCount        *uint16              `json:"pax_count"`
	EventChoice     *string              `json:"event_choice"`
	CheckedInEvents []CheckinEventStatus `json:"checked_in_events"`
}

// CheckinGuest is the guest block of a successful check-in response.
type CheckinGuest struct {
	Name          string  `json:"name"`
	EventType     string  `json:"event_type"`
	RegisteredPax *uint16 `json:"registered_pax"`
	CheckedInPax  uint16  `json:"checked_in_pax"`
}
