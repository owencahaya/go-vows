# go-vows

# Project Vows

A wedding invitation backend system using WhatsApp.

Couples import their guest list via CSV. Guests RSVP through WhatsApp; confirmed
guests receive a single QR code that can be scanned for check-in at the Holy
Matrimony and/or Reception (depending on the guest's event choice).

Built with **Go + Gin + GORM + MySQL**.

> The Meta WhatsApp Cloud API and Meta Media API calls are **stubbed** behind
> interfaces (see `internal/services/whatsapp_service.go` and
> `meta_media_service.go`). They return deterministic fake responses so the full
> pipeline (logging, status transitions) works end-to-end. QR images are
> generated for real using a lightweight PNG encoder. Search for `TODO` to find
> the integration points.

## Tech stack

- Language: Go (1.22+)
- HTTP framework: Gin
- ORM: GORM (MySQL driver)
- Database: MySQL 8
- Config: environment variables (`.env` supported via godotenv)

## Project structure

```
cmd/api/main.go              # entrypoint: wiring, migrate, serve
internal/
  config/                    # env config + DB connection + AutoMigrate
  models/                    # GORM models (events, invitations, whatsapp_logs, checkin_logs)
  repositories/              # data-access layer
  services/                  # business logic + external-service stubs
  handlers/                  # Gin HTTP handlers
  routes/                    # route registration
  dto/                       # request/response structs
  utils/                     # JSON response helpers + token generation
```

## Setup

1. **Clone & install dependencies**

   ```bash
   go mod download
   ```

2. **Create the database**

   ```sql
   CREATE DATABASE vows CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

3. **Configure environment**

   ```bash
   cp .env.example .env
   # edit .env with your DB credentials and Meta WhatsApp Cloud API credentials
   ```

   | Variable               | Description                           | Default                 |
   | ---------------------- | ------------------------------------- | ----------------------- |
   | `APP_PORT`             | HTTP port                             | `8080`                  |
   | `APP_BASE_URL`         | Base URL used inside QR check-in URLs | `http://localhost:8080` |
   | `DB_HOST`              | MySQL host                            | `localhost`             |
   | `DB_PORT`              | MySQL port                            | `3306`                  |
   | `DB_NAME`              | Database name                         | `vows`                  |
   | `DB_USER`              | MySQL user                            | `root`                  |
   | `DB_PASSWORD`          | MySQL password                        | `password`              |
   | `META_ACCESS_TOKEN`    | Meta WhatsApp Cloud API access token  | _empty_                 |
   | `META_PHONE_NUMBER_ID` | Meta phone number id                  | _empty_                 |
   | `META_VERIFY_TOKEN`    | Webhook verification token            | _empty_                 |
   | `META_API_VERSION`     | Graph API version                     | `v23.0`                 |

## Migrations & running

Migrations run automatically via GORM `AutoMigrate` on startup — no separate
migration step needed.

```bash
go run ./cmd/api
```

The server logs `Project Vows API listening on :8080`. Health check:

```bash
curl http://localhost:8080/health
```

To run the project

```bash
go build -o bin/vows-api ./cmd/api
./bin/vows-api
```

## Data model

| Table           | Purpose                                                     |
| --------------- | ----------------------------------------------------------- |
| `events`        | Wedding/couple data (unique `tag`)                          |
| `invitations`   | Guest, RSVP, pax, event choice, QR metadata                 |
| `whatsapp_logs` | Inbound/outbound WhatsApp logs + Meta responses/errors      |
| `checkin_logs`  | QR check-in records, one per `invitation_id` + `event_type` |

Key constraints:

- `invitations`: unique `event_id` + `whatsapp_number` (one number can be
  invited to multiple weddings, but only once per wedding).
- `checkin_logs`: unique `invitation_id` + `event_type` (a guest checks in once
  per event; `both` allows one Holy Matrimony + one Reception check-in).

## CSV format

```csv
tag,guest_name,whatsapp_number
stanley-arum,Budi Santoso,6281234567890
stanley-arum,Sinta Wijaya,6289876543210
kevin-michelle,Budi Santoso,6281234567890
```

A sample file is provided at [`sample_guests.csv`](sample_guests.csv). The event
identified by `tag` must already exist; rows with unknown tags or duplicate
guests are reported in the import summary rather than aborting the import.

## API reference

All responses use the envelope:

```json
{ "status": "success|error", "message": "...", "data": {...}, "error": "code" }
```

### Events

**Create event** — `POST /api/events`

```bash
curl -X POST http://localhost:8080/api/events \
  -H 'Content-Type: application/json' \
  -d '{
    "tag": "stanley-arum",
    "couple_name": "Stanley & Arum",
    "holy_matrimony_date": "2026-06-20T10:00:00+07:00",
    "holy_matrimony_location": "Gereja ABC",
    "reception_date": "2026-06-20T18:00:00+07:00",
    "reception_location": "Ballroom XYZ",
    "gift_address": "Alamat hadiah",
    "bank_account": "BCA 123456789 a/n Stanley"
  }'
```

- `GET /api/events` — list events
- `GET /api/events/:id` — event detail

### Invitations

**Import CSV** — `POST /api/invitations/import-csv` (multipart, field `file`)

```bash
curl -X POST http://localhost:8080/api/invitations/import-csv \
  -F 'file=@sample_guests.csv'
```

Returns an import summary: `total_rows`, `success_count`, `failed_count`,
`failed_rows[]` (with row number + reason).

**List invitations** — `GET /api/invitations`

Query params: `tag`, `event_id`, `rsvp_status`, `invitation_status`, `qr_status`.

```bash
curl 'http://localhost:8080/api/invitations?tag=stanley-arum&rsvp_status=attending'
```

- `GET /api/invitations/:id` — invitation detail

### WhatsApp send (stubbed)

| Endpoint                                  | Body                      | Targets                                         |
| ----------------------------------------- | ------------------------- | ----------------------------------------------- |
| `POST /api/invitations/send`              | `{"ids":[1,2,3]}`         | the listed invitations                          |
| `POST /api/invitations/send-pending`      | `{"tag":"stanley-arum"}`  | `invitation_status = imported`                  |
| `POST /api/invitations/resend-unanswered` | `{"tag":"stanley-arum"}`  | `rsvp_status = not_answered`                    |
| `POST /api/invitations/send-reminder`     | `{"tag":"stanley-arum"}`  | `rsvp_status = attending`                       |
| `POST /api/invitations/generate-send-qr`  | `{"tag":"stanley-arum"}`  | attending + pax set + choice set + not yet sent |
| `POST /api/invitations/resend-qr`         | `{"tag":"...","ids":[1]}` | selected ids (regenerates media if missing)     |

All of these write rows into `whatsapp_logs` and update the relevant status
fields. Example:

```bash
curl -X POST http://localhost:8080/api/invitations/send \
  -H 'Content-Type: application/json' -d '{"ids":[1,2,3]}'
```

### WhatsApp webhook

**Verify** — `GET /api/webhook/whatsapp`

Implements the Meta handshake: returns `hub.challenge` when
`hub.verify_token` equals `META_VERIFY_TOKEN`.

```bash
curl 'http://localhost:8080/api/webhook/whatsapp?hub.mode=subscribe&hub.verify_token=local-verify-token&hub.challenge=12345'
# -> 12345
```

**Receive** — `POST /api/webhook/whatsapp`

Accepts inbound events and returns `200 OK`. Parsing of Meta interactive button
replies is left as a `TODO`; the conversation mutators (update RSVP, pax, event
choice, gift interest, trigger next message, trigger QR) are implemented in
`internal/services/webhook_service.go` ready to be wired up.

### Check-in

**Lookup** — `GET /api/check-in/:qr_code_token`

```json
{
  "guest_name": "Budi Santoso",
  "pax_count": 2,
  "event_choice": "both",
  "checked_in_events": [
    { "event_type": "holy_matrimony", "checked_in_at": null },
    { "event_type": "reception", "checked_in_at": null }
  ]
}
```

**Check-in** — `POST /api/check-in`

```bash
curl -X POST http://localhost:8080/api/check-in \
  -H 'Content-Type: application/json' \
  -d '{
    "qr_code_token": "8f9d1b2c-a123-4ef0-a999",
    "event_type": "reception",
    "actual_pax": 2,
    "scanner_name": "Admin 1"
  }'
```

Validation:

- invitation must exist (`invitation_not_found`)
- `rsvp_status` must be `attending` (`not_attending`)
- `event_choice` must allow the requested `event_type` (`invalid_event`)
- `actual_pax` must be `>= 1` and `<= pax_count` (`invalid_pax`)
- one check-in per `event_type` (`already_checked_in`)

Success response:

```json
{
  "status": "success",
  "message": "Check-in berhasil",
  "data": {
    "name": "Budi Santoso",
    "event_type": "reception",
    "registered_pax": 2,
    "checked_in_pax": 2
  }
}
```

## Notes / future work

- Replace the WhatsApp and Meta Media stubs with real Graph API HTTP calls.
- Implement Meta webhook payload parsing and the RSVP conversation state machine.
- The QR currently encodes `APP_BASE_URL/check-in/<qr_code_token>`.
