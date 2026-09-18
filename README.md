# ai-sdlc-backend

Backend source code for the car management system, implemented in Go with a
PostgreSQL database. This repository contains the database schema, the
tooling used to apply it, and a REST API server exposing the vehicle
lifecycle status transition endpoints.

## Technology Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate) with the `pgx/v5` driver
- **API style:** REST (no ORM)
- **Authentication:** JWT bearer tokens (HS256+), validated per the [technical requirement documents](https://github.com/mshahkap33/ai-sdlc/tree/main/docs/trd)

## Repository Layout

```
cmd/migrate/           CLI entrypoint that applies or rolls back migrations
cmd/server/            REST API server entrypoint
internal/auth/         JWT bearer-token authentication and role-based authorization middleware
internal/config/       Loads database/server connection settings from environment variables (or a .env file)
internal/db/           Migration runner built on golang-migrate (Migrate / Rollback)
internal/vehiclestatus/  Vehicle lifecycle status state machine, service, repository, and HTTP handlers
db/migrations/         Versioned SQL migration files (one table per file, up/down pairs)
docs/exaples/          Usage documentation (e.g. curl samples) for the REST APIs
```

## Database Schema

The schema covers the car management domain as defined in the technical
requirement documents: vehicle onboarding, vehicle-type booking, vehicle
status/availability, and maintenance/service scheduling.

Migrations are applied in order from `db/migrations/`:

| # | Migration | Description |
| --- | --- | --- |
| 000001 | `enable_pgcrypto_extension` | Enables the `pgcrypto` extension so `gen_random_uuid()` can be used as the default for UUID primary keys |
| 000002 | `create_vehicle_categories_table` | Vehicle categories/classes (e.g. Economy, SUV, Luxury) |
| 000003 | `create_vehicles_table` | Canonical vehicle record (identity, condition, and lifecycle `status`) |
| 000004 | `create_booking_policies_table` | Per-category rules controlling allowed reservation modes |
| 000005 | `create_reservations_table` | Customer reservations, at the vehicle or category level |
| 000006 | `create_reservation_vehicle_assignments_table` | The specific vehicle assigned to fulfill a reservation |
| 000007 | `create_vehicle_status_history_table` | Append-only audit trail of vehicle status transitions |
| 000008 | `create_vehicle_compliance_documents_table` | Registration, insurance, inspection, and recall records |
| 000009 | `create_vehicle_status_triggers_table` | Configurable mapping of events to automatic status transitions |
| 000010 | `create_maintenance_schedules_table` | Maintenance-trigger rules, per category or per vehicle |
| 000011 | `create_vehicle_telematics_readings_table` | Raw telematics readings (mileage, diagnostics) used to evaluate schedules |
| 000012 | `create_maintenance_work_orders_table` | Work orders generated from schedules, recalls, or safety defects |
| 000013 | `create_maintenance_alerts_table` | Alerts raised for overdue service, expiring documents, and recalls |

Each table migration includes its indexes and constraints (foreign keys,
`CHECK` constraints for enum-like fields, and uniqueness constraints) in the
same SQL file, following a one-table-per-file convention. Every table also
carries the standard audit/soft-delete columns: `created_at`, `updated_at`,
`deleted_at`, `created_by`, `updated_by`, and `deleted`.

## Configuration

Connection settings are read from environment variables (see
`internal/config/config.go` for defaults). A root-level [`.env`](./.env) file
with these defaults is loaded automatically on startup via
[godotenv](https://github.com/joho/godotenv); real environment variables
always take precedence over the values in `.env`, so it is safe to override
any of them (e.g. in CI or production) without editing the file. For local,
untracked overrides, copy values into a `.env.local` file, which is ignored
by git.

| Variable | Default |
| --- | --- |
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_USER` | `postgres` |
| `DB_PASSWORD` | `postgres` |
| `DB_NAME` | `ai_sdlc` |
| `DB_SSLMODE` | `disable` |
| `SERVER_ADDR` | `:8080` |
| `JWT_SECRET` | *(none, must be set to run the server)* |

## Running Migrations

Apply all pending migrations:

```sh
go run ./cmd/migrate -direction up
```

Roll back the last migration (use `-steps` to roll back more than one):

```sh
go run ./cmd/migrate -direction down -steps 1
```

By default the CLI reads migration files from `db/migrations`; override the
location with `-path` if needed.

## REST API

Start the server (requires `JWT_SECRET` to be set):

```sh
go run ./cmd/server
```

This currently exposes the vehicle lifecycle status transition endpoints
defined in the [Vehicle Status and Availability TRD](https://github.com/mshahkap33/ai-sdlc/blob/main/docs/trd/trd-vehicle-status-availability.md):

| Endpoint | Description | Required role |
| --- | --- | --- |
| `POST /api/v1/vehicles/{vehicleId}/status-events` | Submit a triggering event (e.g. `booking_confirmed`) that automatically transitions the vehicle's status | `system_service` |
| `PATCH /api/v1/vehicles/{vehicleId}/status` | Manually override a vehicle's status with a mandatory reason | `service_staff` or `operations_manager` |

Every request must include a JWT in the `Authorization` header whose
`roles` claim contains one of the required roles above. See
[`docs/exaples/transition-vehicle-lifecycle-status.md`](./docs/exaples/transition-vehicle-lifecycle-status.md)
for curl examples and error responses.

## Testing

```sh
go test ./...
```

The migration tests in `internal/db` apply and roll back the full schema
against a real PostgreSQL instance. They are skipped automatically unless the
`TEST_DATABASE_DSN` environment variable is set, e.g.:

```sh
TEST_DATABASE_DSN="host=localhost port=5432 user=postgres dbname=ai_sdlc sslmode=disable" go test ./internal/db/...
```
