# Hotel Booking Microservices Platform

Backend-only hotel booking system built with Go 1.21, PostgreSQL, Docker, and a Domain-Driven Design structure. The platform is decomposed into six microservices behind an API Gateway and demonstrates SOLID principles, clean layering, payment-gateway abstraction, graceful shutdowns, structured logging, and automated Swagger documentation.

---

## Architecture Overview

```
                    +--------------------+
Request             |  API Gateway (8088)|--- reverse proxy --> individual services
Bearer JWT          |  - JWT middleware  |--- aggregation --> Booking + Payment
------------------->|  - Rate limiting   |
                    +--------------------+
                              |
    --------------------------------------------------------------------------
    |            |              |                 |                  |        |
+--------+  +-----------+  +-----------+    +------------+    +--------------+|
| Auth   |  | Hotel     |  | Booking   |    | Payment    |    | Notification ||
| 8080   |  | 8081      |  | 8082      |    | 8083       |    | 8085         ||
| JWT &  |  | Hotels &  |  | Booking   |    | Payment    |    | Logger-based ||
| users  |  | rooms     |  | lifecycle |    | provider   |    | dispatcher   ||
+--------+  +-----------+  +-----------+    +------------+    +--------------+|
    |             |              |                 |                  |
    ----------------------------- PostgreSQL (5432) ---------------------------
```

**Bounded contexts & responsibilities**

| Service           | Responsibilities                                                                                   |
|-------------------|----------------------------------------------------------------------------------------------------|
| API Gateway       | JWT verification, rate limiting, reverse proxy, booking+payment aggregation                        |
| Auth Service      | Register/login, password hashing (bcrypt), JWT issuing, `/register /login /me`                     |
| Hotel Service     | CRUD hotels, room types, rooms, public listing with room type summaries                             |
| Booking Service   | Booking lifecycle (create → pending → confirmed → checked_in → completed), cancellations, check-ins|
| Payment Service   | PaymentProvider abstraction (mock Xendit), initiation, webhook verification, refunds, booking sync |
| Notification Svc  | Simple dispatcher (zap logger), triggered on booking creation or payment events                    |

Shared packages live under `pkg/` (config, logger, middleware, dto, etc.). Each service follows DDD layers: `internal/domain`, `internal/usecase`, `internal/infrastructure`.

---

## Tech Stack & Key Features

- **Language**: Go 1.21
- **Frameworks**: chi router, sqlx, jwt v5, zap logger
- **Database**: PostgreSQL with UUIDs via `uuid-ossp`
- **Containerization**: Docker + docker-compose (multi-stage builds)
- **Docs**: Swagger generated from handler annotations (`docs/swagger/swagger.yaml`)
- **Testing**: testify, sqlmock; table-driven tests covering booking lifecycle, payment webhooks, repository inserts
- **Non-functional**:
  - JWT-based authentication & role checks
  - Rate limiting middleware on gateway
  - Structured logging + context-aware logging helpers
  - Config via environment variables (`.env.example`)
  - Graceful shutdown using context cancellation & signal handling
  - Payment provider abstraction + mock Xendit signature validation
  - Consistent API error contract (DTO-based)

---

## Repository Layout

```
cmd/<service>/           # each service entry point (auth-service, booking-service, etc.)
internal/
  domain/<bounded>       # entities & repository interfaces
  usecase/<bounded>      # core business logic
  infrastructure/        # http handlers, repositories, provider clients, gateway logic
pkg/                     # shared libs (config, dto, middleware, logger, etc.)
migrations/001_init.sql  # SQL schema seed
build/<service>/Dockerfile
docs/                    # Swagger + ERD
```

---

## Prerequisites

- Docker Desktop / Docker Engine 24+
- Docker Compose V2
- Go 1.21+ (only needed for local dev/tests)
- `swag` CLI for regenerating docs (optional): `go install github.com/swaggo/swag/cmd/swag@latest`

---

## Configuration

Copy `.env.example` to `.env` (or supply env vars in compose):

| Variable                     | Default                                      | Description                                  |
|-----------------------------|----------------------------------------------|----------------------------------------------|
| `DATABASE_URL`              | `postgres://postgres:postgres@postgres:5432/hotel?sslmode=disable` | Shared Postgres DSN                |
| `JWT_SECRET`                | `super-secret`                               | JWT signing secret                           |
| `PAYMENT_PROVIDER_KEY`      | `sandbox-key`                                | HMAC key for mock Xendit                     |
| `PAYMENT_SERVICE_URL`       | `http://payment-service:8083`                | Booking service -> payment client target     |
| `BOOKING_SERVICE_URL`       | `http://booking-service:8082`                | Payment service -> booking status callback   |
| `NOTIFICATION_SERVICE_URL`  | `http://notification-service:8085`           | Booking service -> notification client       |
| `AGGREGATE_TARGET_URL`      | `http://hotel-service:8081`                  | Gateway reverse proxy target for `/proxy`    |
| `RATE_LIMIT_PER_MINUTE`     | `120`                                        | Gateway rate limiter                         |

Each service also respects `HTTP_PORT` env for its listener.

---

## Running the Stack

1. **Spin up services**
   ```bash
   docker-compose up --build
   ```
   - API Gateway: `http://localhost:8088/gateway`
   - Auth: `http://localhost:8080`
   - Hotel: `http://localhost:8081`
   - Booking: `http://localhost:8082`
   - Payment: `http://localhost:8083`
   - Notification: `http://localhost:8085`
   - PostgreSQL: `localhost:5432`
   - Adminer UI: `http://localhost:8089`

2. **Shutdown & cleanup**
   ```bash
   docker-compose down -v
   ```

---

## Database & Migrations

- `migrations/001_init.sql` creates tables:
  - `users`, `hotels`, `room_types`, `rooms`, `bookings`, `payments`, `refunds`, `checkins`
  - All keys use UUID (requires `CREATE EXTENSION "uuid-ossp"`).
  - Indexes on FK columns for performant joins.
- Run the SQL manually (psql into the Postgres container) or hook up a migration tool (goose, migrate, etc.). Docker compose seeds nothing by default.

**ERD Snapshot**

| Entity     | Highlights                              | Relationships                                            |
|------------|-----------------------------------------|---------------------------------------------------------|
| users      | email, hashed password, role            | 1 - * bookings                                           |
| hotels     | name, address                           | 1 - * room_types                                         |
| room_types | hotel_id, capacity, base_price          | 1 - * rooms, * - 1 hotels                                |
| rooms      | room_type_id, number, status            | * - 1 room_types                                         |
| bookings   | user_id, room_type_id, status, totals   | * - 1 users, * - 1 room_types, 1 - 1 payments/checkins   |
| payments   | booking_id, provider, amount            | 1 - 1 bookings, 1 - * refunds                            |
| refunds    | payment_id, status                      | * - 1 payments                                           |
| checkins   | booking_id, timestamps                  | 1 - 1 bookings                                           |

---

## Swagger / API Documentation

1. **Generate docs**
   ```bash
   make swagger
   ```
   Produces `docs/swagger/swagger.yaml` & `docs/swagger/swagger.json`.

2. **Serve via swagger-ui** (optional)
   ```bash
   docker run --rm -p 8087:8080 \
     -e SWAGGER_JSON=/app/swagger.yaml \
     -v ${PWD}/docs/swagger/swagger.yaml:/app/swagger.yaml \
     swaggerapi/swagger-ui
   ```
   Visit `http://localhost:8087`.

3. **Postman import**  
   - Import `docs/swagger/swagger.yaml` directly or consume `docs/swagger/swagger.json`.

Swagger covers:
- Auth `/register /login /me/{id}`
- Hotel `/hotels /room-types /rooms`
- Booking `/bookings`, cancellation, checkpoint, internal status sync
- Payment `/payments`, `/payments/{id}`, `/payments/webhook`, `/payments/refund`
- Notification `/notifications`
- Gateway `/gateway/aggregate/bookings/{id}`

Use the Swagger “Authorize” button with `Bearer <access_token>` for protected endpoints.

---

## Testing & Linting

```bash
make test     # go test ./... -cover
make lint     # go vet ./...
```

Covered scenarios include:
- Booking creation: happy path, invalid dates, room type not found.
- Payment webhook: valid/invalid signature propagation with booking status updates.
- Refund flows: provider success/fail.
- Booking repository insert via sqlmock.

Target coverage: ~45–60% concentrating on business logic layers.

---

## Example Workflow (Manual / Postman)

1. **Register admin user**
   ```
   POST http://localhost:8080/register
   {
     "email": "admin@example.com",
     "password": "Secret123!",
     "role": "admin"
   }
   ```
2. **Login** and copy `access_token`.
3. **Create hotel + room types** (Auth header `Bearer <token>`) via Hotel service.
4. **Register customer**, login, and create a booking:
   ```
   POST http://localhost:8082/bookings
   {
     "user_id": "customer-uuid",
     "room_type_id": "roomtype-uuid",
     "check_in": "2025-12-20T15:00:00Z",
     "check_out": "2025-12-23T11:00:00Z",
     "guests": 2
   }
   ```
5. **Initiate payment** via Payment service.
6. **Simulate webhook** (mock) to mark payment paid:
   ```
   POST /payments/webhook
   {
     "payment_id": "...",
     "status": "paid",
     "signature": "<hmac of payload>"
   }
   ```
   Booking status auto-updates to `confirmed`.
7. **Aggregate view**:
   ```
   GET http://localhost:8088/gateway/aggregate/bookings/{booking_id}
   ```

Use Adminer (`http://localhost:8089`) to inspect database tables while testing.

---

## Makefile Cheat Sheet

| Command        | Description                                 |
|----------------|---------------------------------------------|
| `make run`     | `docker-compose up --build`                 |
| `make down`    | `docker-compose down -v`                    |
| `make test`    | `go test ./... -cover`                      |
| `make lint`    | `go vet ./...`                              |
| `make swagger` | Generate Swagger docs under `docs/swagger/` |

---

## Next Steps & Customization

- Swap the mock payment provider with a real Midtrans/Xendit integration by implementing `domain.Provider`.
- Replace notification logger dispatcher with email/SMS providers by extending `domain.Dispatcher`.
- Add caching layers (Redis) or message brokers (NATS/Kafka) if needed.
- Wire CI/CD (GitHub Actions) to run `make test` and `make lint`.
- Add more bounded contexts (inventory, pricing) by following the same DDD folder layout.

---

## License

This repository is intended as a technical test / sample implementation; no explicit license is provided. Adapt and extend as needed for your environment.
