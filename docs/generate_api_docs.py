#!/usr/bin/env python3
"""Generate the Project Vows API documentation PDF for the front-end team."""

from reportlab.lib.pagesizes import A4
from reportlab.lib.units import mm
from reportlab.lib import colors
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.enums import TA_LEFT
from reportlab.platypus import (
    BaseDocTemplate, PageTemplate, Frame, Paragraph, Spacer, Table, TableStyle,
    HRFlowable, KeepTogether, PageBreak,
)
from reportlab.platypus.flowables import Flowable
from xml.sax.saxutils import escape

# ---------------------------------------------------------------- palette
INK      = colors.HexColor("#1f2933")
MUTED    = colors.HexColor("#647382")
ACCENT   = colors.HexColor("#9b1c47")   # deep wine — wedding theme
ACCENT_L = colors.HexColor("#f6e7ee")
CODE_BG  = colors.HexColor("#1e293b")
CODE_FG  = colors.HexColor("#e2e8f0")
LINE     = colors.HexColor("#e2e8f0")
TABLE_HD = colors.HexColor("#f1f5f9")

METHOD_COLORS = {
    "GET":    colors.HexColor("#1d6fb8"),
    "POST":   colors.HexColor("#1f8a4c"),
    "PUT":    colors.HexColor("#b8791d"),
    "DELETE": colors.HexColor("#b8341d"),
}

# ---------------------------------------------------------------- styles
styles = getSampleStyleSheet()

def S(name, **kw):
    styles.add(ParagraphStyle(name, **kw))

S("DocTitle", fontName="Helvetica-Bold", fontSize=30, leading=34, textColor=ACCENT)
S("DocSub", fontName="Helvetica", fontSize=13, leading=18, textColor=MUTED, spaceBefore=6)
S("H1", fontName="Helvetica-Bold", fontSize=18, leading=22, textColor=INK,
  spaceBefore=18, spaceAfter=6)
S("H2", fontName="Helvetica-Bold", fontSize=13.5, leading=18, textColor=ACCENT,
  spaceBefore=14, spaceAfter=4)
S("Body", fontName="Helvetica", fontSize=10, leading=15, textColor=INK, spaceAfter=6)
S("Small", fontName="Helvetica", fontSize=8.5, leading=12, textColor=MUTED)
S("Path", fontName="Courier-Bold", fontSize=11.5, leading=14, textColor=INK)
S("CellH", fontName="Helvetica-Bold", fontSize=9, leading=12, textColor=INK)
S("Cell", fontName="Helvetica", fontSize=9, leading=12, textColor=INK)
S("CellCode", fontName="Courier", fontSize=8.5, leading=11, textColor=INK)
S("CellMuted", fontName="Helvetica", fontSize=9, leading=12, textColor=MUTED)
S("CodeBlk", fontName="Courier", fontSize=8.3, leading=12, textColor=CODE_FG)
S("Label", fontName="Helvetica-Bold", fontSize=8.5, leading=11, textColor=MUTED,
  spaceBefore=6, spaceAfter=2)
S("TOC", fontName="Helvetica", fontSize=10.5, leading=20, textColor=INK)


# ---------------------------------------------------------------- flowables
class MethodBadge(Flowable):
    """A coloured HTTP-method pill followed by the route path."""
    def __init__(self, method, path):
        super().__init__()
        self.method = method
        self.path = path
        self.height = 18

    def wrap(self, availW, availH):
        self.width = availW
        return availW, self.height

    def draw(self):
        c = self.canv
        m = self.method
        bg = METHOD_COLORS.get(m, MUTED)
        c.setFont("Helvetica-Bold", 9)
        pill_w = c.stringWidth(m, "Helvetica-Bold", 9) + 14
        c.setFillColor(bg)
        c.roundRect(0, 0, pill_w, 15, 3, fill=1, stroke=0)
        c.setFillColor(colors.white)
        c.drawString(7, 4, m)
        c.setFillColor(INK)
        c.setFont("Courier-Bold", 11)
        c.drawString(pill_w + 8, 3, self.path)


def code_block(text):
    """A dark, rounded code block. Returns a single-cell Table flowable."""
    safe = escape(text)
    p = Paragraph(safe.replace("\n", "<br/>").replace(" ", "&nbsp;"), styles["CodeBlk"])
    t = Table([[p]], colWidths=[170 * mm])
    t.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, -1), CODE_BG),
        ("LEFTPADDING", (0, 0), (-1, -1), 10),
        ("RIGHTPADDING", (0, 0), (-1, -1), 10),
        ("TOPPADDING", (0, 0), (-1, -1), 8),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 8),
        ("ROUNDEDCORNERS", [4, 4, 4, 4]),
    ]))
    return t


def param_table(rows, headers=("Field", "Type", "Required", "Description")):
    data = [[Paragraph(h, styles["CellH"]) for h in headers]]
    for r in rows:
        cells = []
        for i, val in enumerate(r):
            style = "CellCode" if i == 0 else ("Cell" if i != len(r) - 1 else "CellMuted")
            cells.append(Paragraph(str(val), styles[style]))
        data.append(cells)
    widths = [34 * mm, 26 * mm, 22 * mm, 88 * mm][:len(headers)]
    if len(headers) == 2:
        widths = [40 * mm, 130 * mm]
    t = Table(data, colWidths=widths, repeatRows=1)
    t.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), TABLE_HD),
        ("LINEBELOW", (0, 0), (-1, 0), 0.6, MUTED),
        ("LINEBELOW", (0, 1), (-1, -1), 0.4, LINE),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("LEFTPADDING", (0, 0), (-1, -1), 6),
        ("RIGHTPADDING", (0, 0), (-1, -1), 6),
        ("TOPPADDING", (0, 0), (-1, -1), 5),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
    ]))
    return t


def rule():
    return HRFlowable(width="100%", thickness=0.6, color=LINE,
                      spaceBefore=10, spaceAfter=10)


# ---------------------------------------------------------------- content
story = []


def endpoint(num, title, method, path, desc, *, query=None, body=None,
             body_rows=None, req_example=None, resp_example=None,
             errors=None, notes=None):
    block = []
    block.append(Paragraph(f"{num}. {title}", styles["H2"]))
    block.append(MethodBadge(method, path))
    block.append(Spacer(1, 6))
    block.append(Paragraph(desc, styles["Body"]))
    if query:
        block.append(Paragraph("QUERY PARAMETERS", styles["Label"]))
        block.append(param_table(query))
    if body_rows:
        block.append(Paragraph("REQUEST BODY", styles["Label"]))
        block.append(param_table(body_rows))
    if req_example:
        block.append(Paragraph("REQUEST EXAMPLE", styles["Label"]))
        block.append(code_block(req_example))
    if resp_example:
        block.append(Paragraph("RESPONSE EXAMPLE", styles["Label"]))
        block.append(code_block(resp_example))
    if errors:
        block.append(Paragraph("POSSIBLE ERRORS", styles["Label"]))
        block.append(param_table(errors, headers=("Error code", "When it happens")))
    if notes:
        block.append(Spacer(1, 4))
        block.append(Paragraph(notes, styles["Small"]))
    block.append(rule())
    # Keep the header + badge with the first piece of content.
    story.append(KeepTogether(block[:4]))
    story.extend(block[4:])


# ===== Cover ====================================================
story.append(Spacer(1, 40 * mm))
story.append(Paragraph("Project Vows", styles["DocTitle"]))
story.append(Paragraph("REST API Documentation", styles["DocSub"]))
story.append(Spacer(1, 4))
story.append(Paragraph("Wedding invitation &amp; WhatsApp RSVP backend", styles["DocSub"]))
story.append(Spacer(1, 18))
story.append(HRFlowable(width="40%", thickness=1.5, color=ACCENT, hAlign="LEFT"))
story.append(Spacer(1, 14))
story.append(Paragraph("Prepared for the Front-End Team", styles["Body"]))
story.append(Paragraph("Version 1.0 &nbsp;&middot;&nbsp; May 2026", styles["Small"]))
story.append(PageBreak())

# ===== Overview =================================================
story.append(Paragraph("Overview", styles["H1"]))
story.append(Paragraph(
    "Project Vows is the backend for a WhatsApp-based wedding invitation system. "
    "Couples (events) import a guest list via CSV; guests RSVP over WhatsApp and "
    "confirmed guests receive a QR code used for check-in at the Holy Matrimony "
    "and/or Reception. This document describes every HTTP endpoint the front end "
    "will call.", styles["Body"]))

story.append(Paragraph("Base URL", styles["H2"]))
story.append(code_block("http://localhost:8080        # local development\n"
                        "https://vows.id              # production (example)"))
story.append(Paragraph("All endpoints below are prefixed with <b>/api</b> unless noted "
                       "(the health check lives at the root).", styles["Body"]))

story.append(Paragraph("Response envelope", styles["H2"]))
story.append(Paragraph(
    "Every JSON response uses the same envelope. <b>status</b> is always present; "
    "the other fields appear depending on the outcome.", styles["Body"]))
story.append(code_block(
    '{\n'
    '  "status":  "success" | "error",\n'
    '  "message": "human readable message",   // optional\n'
    '  "data":    { ... } | [ ... ],          // present on success\n'
    '  "error":   "machine_readable_code"     // present on error\n'
    '}'))
story.append(Paragraph(
    "<b>Front-end tip:</b> branch on the top-level <font face='Courier'>status</font> "
    "field, then read <font face='Courier'>data</font> on success or "
    "<font face='Courier'>error</font> on failure. The <font face='Courier'>error</font> "
    "code is stable and safe to switch on; <font face='Courier'>message</font> is for "
    "display/debugging and may change.", styles["Body"]))

story.append(Paragraph("HTTP status codes", styles["H2"]))
story.append(param_table([
    ["200 OK", "Successful read or processed batch operation"],
    ["201 Created", "Resource created (POST /api/events)"],
    ["400 Bad Request", "Malformed body / failed validation / bad CSV"],
    ["403 Forbidden", "Webhook verification token mismatch"],
    ["404 Not Found", "Event / invitation / QR token not found"],
    ["409 Conflict", "Duplicate (e.g. already checked in, duplicate event tag)"],
    ["422 Unprocessable", "Check-in business-rule failure (not attending, etc.)"],
    ["500 Server Error", "Unexpected server error"],
], headers=("Status", "Meaning")))

story.append(PageBreak())

# ===== Enums ====================================================
story.append(Paragraph("Field reference (enums)", styles["H1"]))
story.append(Paragraph(
    "These string values appear on the invitation object and in filters. Treat them "
    "as fixed string constants.", styles["Body"]))

story.append(Paragraph("invitation_status", styles["H2"]))
story.append(param_table([
    ["imported", "Default. Guest imported from CSV, invitation not yet sent"],
    ["invitation_sent", "Initial invitation message sent successfully"],
    ["invitation_failed", "Send attempt failed"],
], headers=("Value", "Meaning")))

story.append(Paragraph("rsvp_status", styles["H2"]))
story.append(param_table([
    ["not_answered", "Default. Guest has not responded yet"],
    ["attending", "Guest confirmed attendance"],
    ["not_attending", "Guest declined"],
], headers=("Value", "Meaning")))

story.append(Paragraph("event_choice  (nullable)", styles["H2"]))
story.append(param_table([
    ["null", "Not chosen yet"],
    ["holy_matrimony", "Will attend Holy Matrimony only"],
    ["reception", "Will attend Reception only"],
    ["both", "Will attend both events"],
], headers=("Value", "Meaning")))

story.append(Paragraph("qr_status", styles["H2"]))
story.append(param_table([
    ["not_generated", "Default. No QR generated yet"],
    ["generated", "QR generated &amp; uploaded to WhatsApp media"],
    ["sent", "QR delivered to the guest"],
    ["send_failed", "QR send attempt failed"],
], headers=("Value", "Meaning")))

story.append(Paragraph("gift_interest", styles["H2"]))
story.append(param_table([
    ["not_asked", "Default"],
    ["yes", "Guest wants gift / bank details"],
    ["no", "Guest not interested"],
], headers=("Value", "Meaning")))

story.append(Paragraph("event_type  (check-in)", styles["H2"]))
story.append(param_table([
    ["holy_matrimony", "Check-in at the Holy Matrimony"],
    ["reception", "Check-in at the Reception"],
], headers=("Value", "Meaning")))

story.append(PageBreak())

# ===== Events ===================================================
story.append(Paragraph("Events", styles["H1"]))

endpoint(
    "1", "Create event", "POST", "/api/events",
    "Create a wedding/couple record. The <b>tag</b> must be unique and is later "
    "used as the key for CSV import and all batch operations.",
    body_rows=[
        ["tag", "string", "yes", "Unique slug, e.g. \"stanley-arum\""],
        ["couple_name", "string", "yes", "Display name, e.g. \"Stanley &amp; Arum\""],
        ["holy_matrimony_date", "datetime", "no", "RFC3339, e.g. 2026-06-20T10:00:00+07:00"],
        ["holy_matrimony_location", "string", "no", "Free text"],
        ["reception_date", "datetime", "no", "RFC3339"],
        ["reception_location", "string", "no", "Free text"],
        ["gift_address", "string", "no", "Free text"],
        ["bank_account", "string", "no", "Free text"],
    ],
    req_example=(
        '{\n'
        '  "tag": "stanley-arum",\n'
        '  "couple_name": "Stanley & Arum",\n'
        '  "holy_matrimony_date": "2026-06-20T10:00:00+07:00",\n'
        '  "holy_matrimony_location": "Gereja ABC",\n'
        '  "reception_date": "2026-06-20T18:00:00+07:00",\n'
        '  "reception_location": "Ballroom XYZ",\n'
        '  "gift_address": "Alamat hadiah",\n'
        '  "bank_account": "BCA 123456789 a/n Stanley"\n'
        '}'),
    resp_example=(
        '// 201 Created\n'
        '{\n'
        '  "status": "success",\n'
        '  "message": "Event created",\n'
        '  "data": {\n'
        '    "id": 1,\n'
        '    "tag": "stanley-arum",\n'
        '    "couple_name": "Stanley & Arum",\n'
        '    "holy_matrimony_date": "2026-06-20T10:00:00+07:00",\n'
        '    "reception_date": "2026-06-20T18:00:00+07:00",\n'
        '    "created_at": "2026-05-30T16:56:08+07:00",\n'
        '    "updated_at": "2026-05-30T16:56:08+07:00"\n'
        '  }\n'
        '}'),
    errors=[
        ["bad_request", "Missing tag or couple_name, or invalid JSON"],
        ["create_failed", "Duplicate tag (already exists)"],
    ],
)

endpoint(
    "2", "List events", "GET", "/api/events",
    "Return all events, newest first.",
    resp_example=(
        '{\n'
        '  "status": "success",\n'
        '  "data": [ { "id": 1, "tag": "stanley-arum", ... } ]\n'
        '}'),
)

endpoint(
    "3", "Get event detail", "GET", "/api/events/:id",
    "Return a single event by numeric id.",
    resp_example=(
        '{ "status": "success", "data": { "id": 1, "tag": "stanley-arum", ... } }'),
    errors=[["event_not_found", "No event with that id"]],
)

# ===== Invitations ==============================================
story.append(Paragraph("Invitations", styles["H1"]))

endpoint(
    "4", "Import guests (CSV)", "POST", "/api/invitations/import-csv",
    "Upload a CSV of guests. Sent as <b>multipart/form-data</b> with a single file "
    "field named <font face='Courier'>file</font>. The event referenced by each row's "
    "<b>tag</b> must already exist. Each row is processed independently &mdash; bad rows "
    "are reported, they do not abort the import. An <font face='Courier'>invitation_code</font> "
    "and <font face='Courier'>qr_code_token</font> are generated automatically.",
    body_rows=[
        ["file", "file", "yes", "CSV with header: tag,guest_name,whatsapp_number"],
    ],
    req_example=(
        '# multipart/form-data, field "file"\n'
        'tag,guest_name,whatsapp_number\n'
        'stanley-arum,Budi Santoso,6281234567890\n'
        'stanley-arum,Sinta Wijaya,6289876543210'),
    resp_example=(
        '{\n'
        '  "status": "success",\n'
        '  "message": "Import completed",\n'
        '  "data": {\n'
        '    "total_rows": 3,\n'
        '    "success_count": 2,\n'
        '    "failed_count": 1,\n'
        '    "failed_rows": [\n'
        '      { "row": 4, "reason": "event not found for tag",\n'
        '        "tag": "kevin-michelle", "guest_name": "Budi Santoso" }\n'
        '    ]\n'
        '  }\n'
        '}'),
    errors=[
        ["bad_request", "Missing file field, or empty / unreadable CSV"],
    ],
    notes="Note: HTTP 200 is returned even when some rows fail; inspect "
          "failed_count / failed_rows to surface per-row problems to the user. "
          "A row fails on: missing column, missing field, unknown tag, or a "
          "duplicate guest (same whatsapp_number already in that event).",
)

endpoint(
    "5", "List invitations", "GET", "/api/invitations",
    "Return invitations, newest first. All query params are optional and combine "
    "with AND. Each item embeds its parent <font face='Courier'>event</font> object.",
    query=[
        ["tag", "string", "no", "Filter by the event's tag"],
        ["event_id", "number", "no", "Filter by numeric event id"],
        ["rsvp_status", "string", "no", "not_answered | attending | not_attending"],
        ["invitation_status", "string", "no", "imported | invitation_sent | invitation_failed"],
        ["qr_status", "string", "no", "not_generated | generated | sent | send_failed"],
    ],
    req_example="GET /api/invitations?tag=stanley-arum&rsvp_status=attending",
    resp_example=(
        '{\n'
        '  "status": "success",\n'
        '  "data": [\n'
        '    {\n'
        '      "id": 1,\n'
        '      "event_id": 1,\n'
        '      "guest_name": "Budi Santoso",\n'
        '      "whatsapp_number": "6281234567890",\n'
        '      "invitation_code": "inv_7447ed7ae15f97d4",\n'
        '      "invitation_status": "imported",\n'
        '      "rsvp_status": "not_answered",\n'
        '      "pax_count": null,\n'
        '      "event_choice": null,\n'
        '      "gift_interest": "not_asked",\n'
        '      "qr_code_token": "0cc04bab-a900-4e9f-9644-69ce1da41841",\n'
        '      "qr_status": "not_generated",\n'
        '      "qr_sent_at": null,\n'
        '      "created_at": "2026-05-30T16:56:08+07:00",\n'
        '      "event": { "id": 1, "tag": "stanley-arum", ... }\n'
        '    }\n'
        '  ]\n'
        '}'),
)

endpoint(
    "6", "Get invitation detail", "GET", "/api/invitations/:id",
    "Return a single invitation by numeric id, including its event and any "
    "check-in logs.",
    errors=[["invitation_not_found", "No invitation with that id"]],
)

story.append(PageBreak())

# ===== WhatsApp send ============================================
story.append(Paragraph("WhatsApp send operations", styles["H1"]))
story.append(Paragraph(
    "These endpoints trigger outbound WhatsApp messages and update status fields. "
    "They return a <b>batch summary</b> rather than throwing on individual failures, "
    "so a partial failure still returns HTTP 200 &mdash; inspect the per-id "
    "<font face='Courier'>results</font> array.", styles["Body"]))
story.append(Paragraph("Batch summary shape", styles["Label"]))
story.append(code_block(
    '{\n'
    '  "status": "success",\n'
    '  "message": "Batch processed",\n'
    '  "data": {\n'
    '    "total": 2,\n'
    '    "succeeded": 2,\n'
    '    "failed": 0,\n'
    '    "results": [\n'
    '      { "id": 1, "success": true },\n'
    '      { "id": 2, "success": false, "reason": "..." }\n'
    '    ]\n'
    '  }\n'
    '}'))
story.append(rule())

endpoint(
    "7", "Send invitations (by id)", "POST", "/api/invitations/send",
    "Send the initial invitation to a specific set of invitations. Updates "
    "<font face='Courier'>invitation_status</font> to invitation_sent / invitation_failed.",
    body_rows=[["ids", "number[]", "yes", "Invitation ids (at least one)"]],
    req_example='{ "ids": [1, 2, 3] }',
    resp_example='{ "status": "success", "message": "Invitations processed", "data": { ... } }',
    errors=[["bad_request", "Empty or missing ids array"]],
)

endpoint(
    "8", "Send pending invitations", "POST", "/api/invitations/send-pending",
    "Send the initial invitation to <b>all</b> guests of an event whose "
    "<font face='Courier'>invitation_status = imported</font>.",
    body_rows=[["tag", "string", "yes", "Event tag"]],
    req_example='{ "tag": "stanley-arum" }',
)

endpoint(
    "9", "Resend to unanswered", "POST", "/api/invitations/resend-unanswered",
    "Resend the invitation to all guests of an event whose "
    "<font face='Courier'>rsvp_status = not_answered</font>.",
    body_rows=[["tag", "string", "yes", "Event tag"]],
    req_example='{ "tag": "stanley-arum" }',
)

endpoint(
    "10", "Send reminder", "POST", "/api/invitations/send-reminder",
    "Send a reminder to all <font face='Courier'>attending</font> guests of an event.",
    body_rows=[["tag", "string", "yes", "Event tag"]],
    req_example='{ "tag": "stanley-arum" }',
)

endpoint(
    "11", "Generate &amp; send QR", "POST", "/api/invitations/generate-send-qr",
    "Generate, upload and send the QR code to all eligible guests of an event. "
    "A guest is eligible when: rsvp_status = attending, pax_count is set, "
    "event_choice is set, and the QR has not been sent yet. Sets "
    "<font face='Courier'>qr_status = sent</font> and <font face='Courier'>qr_sent_at</font>.",
    body_rows=[["tag", "string", "yes", "Event tag"]],
    req_example='{ "tag": "stanley-arum" }',
    notes="If no guests are eligible, total is 0 and results is empty — this is "
          "still a success response, not an error.",
)

endpoint(
    "12", "Resend QR (by id)", "POST", "/api/invitations/resend-qr",
    "Manually resend the QR to selected invitations within an event. If the QR "
    "media is missing it is regenerated and re-uploaded first. Ids that do not "
    "belong to the given tag are reported as failed in the results array.",
    body_rows=[
        ["tag", "string", "yes", "Event tag (guard)"],
        ["ids", "number[]", "yes", "Invitation ids to resend to"],
    ],
    req_example='{ "tag": "stanley-arum", "ids": [1, 2, 3] }',
)

story.append(PageBreak())

# ===== Webhook ==================================================
story.append(Paragraph("WhatsApp webhook", styles["H1"]))
story.append(Paragraph(
    "These endpoints are called by Meta, not by the front end. They are documented "
    "here for completeness.", styles["Small"]))

endpoint(
    "13", "Verify webhook", "GET", "/api/webhook/whatsapp",
    "Meta verification handshake. Echoes <font face='Courier'>hub.challenge</font> "
    "as plain text when <font face='Courier'>hub.verify_token</font> matches the "
    "server's configured token.",
    query=[
        ["hub.mode", "string", "yes", "Always \"subscribe\""],
        ["hub.verify_token", "string", "yes", "Must match META_VERIFY_TOKEN"],
        ["hub.challenge", "string", "yes", "Echoed back on success"],
    ],
    req_example="GET /api/webhook/whatsapp?hub.mode=subscribe"
                "&hub.verify_token=...&hub.challenge=12345",
    resp_example="200 OK\n12345          # plain text, not JSON",
    errors=[["verification_failed", "Token did not match (403)"]],
)

endpoint(
    "14", "Receive webhook", "POST", "/api/webhook/whatsapp",
    "Receives inbound WhatsApp events from Meta and acknowledges with 200 OK. "
    "Payload parsing of interactive replies is handled server-side.",
    resp_example='{ "status": "received" }',
)

# ===== Check-in =================================================
story.append(Paragraph("Check-in", styles["H1"]))
story.append(Paragraph(
    "These are the endpoints the scanner/check-in UI calls. After scanning a QR, "
    "the encoded URL contains the <font face='Courier'>qr_code_token</font>; use it "
    "to look up the guest, then POST the check-in.", styles["Body"]))

endpoint(
    "15", "Look up guest by QR", "GET", "/api/check-in/:qr_code_token",
    "Resolve a QR token to guest details and per-event check-in status. The "
    "<font face='Courier'>checked_in_events</font> array only includes events the "
    "guest's event_choice covers; <font face='Courier'>checked_in_at</font> is null "
    "until they check in for that event.",
    resp_example=(
        '{\n'
        '  "status": "success",\n'
        '  "data": {\n'
        '    "guest_name": "Budi Santoso",\n'
        '    "pax_count": 2,\n'
        '    "event_choice": "both",\n'
        '    "checked_in_events": [\n'
        '      { "event_type": "holy_matrimony", "checked_in_at": null },\n'
        '      { "event_type": "reception",\n'
        '        "checked_in_at": "2026-05-30T16:56:44+07:00" }\n'
        '    ]\n'
        '  }\n'
        '}'),
    errors=[["invitation_not_found", "No invitation for that QR token (404)"]],
)

endpoint(
    "16", "Check in", "POST", "/api/check-in",
    "Record a check-in for one event. Validated against RSVP status, event choice, "
    "pax count, and duplicate check-in. See the validation rules below.",
    body_rows=[
        ["qr_code_token", "string", "yes", "From the scanned QR"],
        ["event_type", "string", "yes", "holy_matrimony | reception"],
        ["actual_pax", "number", "yes", "People arriving now; 1 .. pax_count"],
        ["scanner_name", "string", "no", "Operator name, for the audit log"],
    ],
    req_example=(
        '{\n'
        '  "qr_code_token": "0cc04bab-a900-4e9f-9644-69ce1da41841",\n'
        '  "event_type": "reception",\n'
        '  "actual_pax": 2,\n'
        '  "scanner_name": "Admin 1"\n'
        '}'),
    resp_example=(
        '// 200 OK\n'
        '{\n'
        '  "status": "success",\n'
        '  "message": "Check-in berhasil",\n'
        '  "data": {\n'
        '    "name": "Budi Santoso",\n'
        '    "event_type": "reception",\n'
        '    "registered_pax": 2,\n'
        '    "checked_in_pax": 2\n'
        '  }\n'
        '}'),
    errors=[
        ["invitation_not_found", "Unknown QR token (404)"],
        ["not_attending", "rsvp_status is not \"attending\" (422)"],
        ["invalid_event", "event_choice does not allow this event_type (422)"],
        ["invalid_pax", "actual_pax is 0 or greater than pax_count (422)"],
        ["already_checked_in", "Guest already checked in for this event_type (409)"],
    ],
)

story.append(Paragraph("Check-in validation rules", styles["H2"]))
story.append(param_table([
    ["1. Token", "Invitation must exist for the qr_code_token, else invalid_not_found"],
    ["2. RSVP", "rsvp_status must be \"attending\""],
    ["3. Event match", "holy_matrimony &rarr; only holy_matrimony; reception &rarr; only "
                       "reception; both &rarr; either"],
    ["4. Pax", "1 &le; actual_pax &le; pax_count"],
    ["5. Uniqueness", "Only one check-in per (invitation, event_type)"],
], headers=("Check", "Rule")))
story.append(Spacer(1, 6))
story.append(Paragraph(
    "<b>Recommended UI flow:</b> scan QR &rarr; GET /api/check-in/{token} to show the "
    "guest and which events are still open &rarr; let the operator pick the event and "
    "enter actual_pax &rarr; POST /api/check-in. Map the returned "
    "<font face='Courier'>error</font> code to a clear on-screen message "
    "(e.g. already_checked_in &rarr; \"Already checked in for this event\").",
    styles["Body"]))

story.append(PageBreak())

# ===== Error code index =========================================
story.append(Paragraph("Error code index", styles["H1"]))
story.append(Paragraph(
    "Complete list of machine-readable <font face='Courier'>error</font> codes the "
    "front end may receive.", styles["Body"]))
story.append(param_table([
    ["bad_request", "400", "Validation / malformed body / bad multipart"],
    ["event_not_found", "404", "GET /api/events/:id with unknown id"],
    ["create_failed", "409", "Duplicate event tag on create"],
    ["invitation_not_found", "404", "Unknown invitation id or QR token"],
    ["verification_failed", "403", "Webhook verify token mismatch"],
    ["not_attending", "422", "Check-in: guest not attending"],
    ["invalid_event", "422", "Check-in: event_choice mismatch"],
    ["invalid_pax", "422", "Check-in: actual_pax out of range"],
    ["already_checked_in", "409", "Check-in: duplicate for event_type"],
    ["internal_error", "500", "Unexpected server error"],
], headers=("Error code", "HTTP", "Meaning")))

story.append(Spacer(1, 16))
story.append(HRFlowable(width="100%", thickness=0.6, color=LINE))
story.append(Spacer(1, 6))
story.append(Paragraph(
    "Project Vows API &middot; v1.0 &middot; generated for the front-end team. "
    "Datetimes are RFC3339 with timezone offset. All request/response bodies are "
    "<font face='Courier'>application/json</font> except CSV import (multipart) and "
    "the webhook verify response (plain text).", styles["Small"]))


# ---------------------------------------------------------------- build
def footer(canvas, doc):
    canvas.saveState()
    canvas.setFont("Helvetica", 8)
    canvas.setFillColor(MUTED)
    canvas.drawString(20 * mm, 12 * mm, "Project Vows — API Documentation")
    canvas.drawRightString(190 * mm, 12 * mm, f"Page {doc.page}")
    canvas.setStrokeColor(LINE)
    canvas.line(20 * mm, 15 * mm, 190 * mm, 15 * mm)
    canvas.restoreState()


doc = BaseDocTemplate(
    "/Users/owencahaya/Documents/go/go-vows/docs/ProjectVows_API_Documentation.pdf",
    pagesize=A4,
    leftMargin=20 * mm, rightMargin=20 * mm,
    topMargin=18 * mm, bottomMargin=20 * mm,
    title="Project Vows API Documentation",
    author="Project Vows",
)
frame = Frame(doc.leftMargin, doc.bottomMargin,
              doc.width, doc.height, id="main")
doc.addPageTemplates([PageTemplate(id="all", frames=[frame], onPage=footer)])
doc.build(story)
print("PDF written")
