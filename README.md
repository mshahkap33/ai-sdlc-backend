# ai-sdlc-backend

Backend source code for the car management system, implemented in Go with a
PostgreSQL database. This repository contains the database schema, the
tooling used to apply it, and the REST API implementing the car management
capabilities.

## Technology Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate) with the `pgx/v5` driver
- **API style:** REST (no ORM)

## Repository Layout

```
cmd/migrate/           CLI entrypoint that applies or rolls back migrations
cmd/server/            HTTP server entrypoint exposing the REST API
internal/auth/         JWT bearer-token authentication and role authorization middleware
internal/config/       Loads database/auth/server settings from environment variables (or a .env file)
internal/db/           Migration runner built on golang-migrate (Migrate / Rollback)
internal/vehicle/      Vehicle Onboarding domain logic, repository, and HTTP handlers
db/migrations/         Versioned SQL migration files (one table per file, up/down pairs)
docs/exaples/          Sample API usage (e.g. curl requests)
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

The API server additionally reads the following variables:

| Variable | Default | Notes |
| --- | --- | --- |
| `SERVER_ADDR` | `:8080` | address the HTTP server listens on |
| `AUTH_JWT_PUBLIC_KEY` | _(none)_ | PEM-encoded RSA public key (or certificate) used to verify bearer JWTs; required |
| `AUTH_JWT_PUBLIC_KEY_PATH` | _(none)_ | path to a file containing the PEM key, used when `AUTH_JWT_PUBLIC_KEY` is not set |
| `AUTH_JWT_ISSUER` | _(none)_ | when set, the required JWT `iss` claim |
| `AUTH_JWT_AUDIENCE` | _(none)_ | when set, the required JWT `aud` claim |

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

## Running the API Server

```sh
go run ./cmd/server
```

The server exposes the Vehicle Onboarding REST API. All endpoints require an
RS256-signed JWT bearer token (see `AUTH_JWT_PUBLIC_KEY` above); creating a
vehicle additionally requires the `service_staff` role. See
[`docs/exaples/create-vehicle.md`](./docs/exaples/create-vehicle.md) for a
worked example, including a sample `curl` request.

## Testing

```sh
go test ./...
```

The migration tests in `internal/db` apply and roll back the full schema
against a real PostgreSQL instance. The repository integration tests in
`internal/vehicle` do the same. Both are skipped automatically unless the
`TEST_DATABASE_DSN` environment variable is set, e.g.:

```sh
TEST_DATABASE_DSN="host=localhost port=5432 user=postgres dbname=ai_sdlc sslmode=disable" go test ./internal/db/... ./internal/vehicle/...
```
