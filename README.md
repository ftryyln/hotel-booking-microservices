# Hotel Booking Microservices (Go)

Production-style backend for hotel booking platform using Go 1.21, PostgreSQL, and DDD-oriented microservices (Auth, Hotel, Booking, Payment, Notification, API Gateway).

## Services

- **Auth Service** – registration/login/JWT issuance.
- **Hotel Service** – CRUD hotels, room types, rooms.
- **Booking Service** – booking lifecycle, payment initiation, notifications.
- **Payment Service** – Midtrans/Xendit mock, webhook, refunds, booking status updates.
- **Notification Service** – simple logger-based dispatch.
- **API Gateway** – JWT verification, rate limiting, reverse proxy, booking+payment aggregation.

## Structure

```
cmd/<service>
internal/
  domain/
  usecase/
  infrastructure/
pkg/
  config, logger, errors, dto, middleware, database, server, utils
migrations/
build/
```

Shared DTOs enforce separation between requests, responses, and domain models. Each service respects DDD layers (domain -> usecase -> infrastructure) and SOLID via repository interfaces + dependency injection.

## Getting Started

```bash
cp .env.example .env # adjust as needed
docker-compose up --build
```

Services exposed via Docker Compose:

- API Gateway: `http://localhost:8088`
- Auth: `:8080`
- Hotel: `:8081`
- Booking: `:8082`
- Payment: `:8083`
- Notification: `:8085`
- Adminer: `:8089`
- Postgres: `:5432`

Stop system:

```bash
docker-compose down -v
```

## Database & Migrations

All tables defined in `migrations/001_init.sql` use UUID primary keys plus indexes for FK columns. Run via your preferred migration tool or manually psql into the container.

## Swagger / API Docs

Swagger annotations live directly inside handlers to keep docs near code. Install swag CLI once:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.3
```

Generate docs (output `docs/swagger/` containing `docs.go`, `swagger.yaml`, `swagger.json`):

```bash
make swagger
```

Import `docs/swagger/swagger.yaml` (or serve via swagger-ui) to inspect endpoints spanning Auth, Hotel, Booking, Payment (init/webhook/refund), Notification, and Gateway.

## Testing

Unit tests cover core booking/payment flows plus repository SQL using sqlmock. Run from repo root:

```bash
make test  # go test ./... -cover
```

Coverage target ~50% for domain-critical logic.

## Makefile Targets

- `make run` – `docker-compose up --build`
- `make down` – `docker-compose down -v`
- `make test` – run Go unit tests
- `make lint` – `go vet ./...`
- `make swagger` – regenerate swagger artifacts

## Non-Functional Features

- JWT middleware + role enforcement (pkg/middleware).
- Structured logging via zap shared singleton.
- Graceful shutdown wired for every service (context cancellation + SIGINT/SIGTERM).
- Config via environment variables with `.env.example` template.
- Rate limiting on API gateway (simple token spacing approach).
- Consistent error response contract (`pkg/errors`, `dto.ErrorResponse`).
- Context-aware HTTP clients for cross-service calls.
- Payment provider abstraction for Midtrans/Xendit compatibility.

## ERD

| Entity | Important Fields | Relationships |
|--------|------------------|---------------|
| users | id, email, role | 1-* bookings |
| hotels | id, name | 1-* room_types |
| room_types | id, hotel_id | 1-* rooms, *-1 hotels |
| rooms | id, room_type_id | *-1 room_types |
| bookings | id, user_id, room_type_id | *-1 users, *-1 room_types, 1-1 payments, 1-1 checkins |
| payments | id, booking_id | 1-1 bookings, 1-* refunds |
| refunds | id, payment_id | *-1 payments |
| checkins | id, booking_id | 1-1 bookings |

## Notes

- Payment service automatically updates booking status via HTTP callback after webhook success/fail.
- Notification service currently logs messages; swap dispatcher for SMTP/SMS/etc.
- Replace mock provider with real Midtrans/Xendit integration by implementing `domain.Provider`.
