# ai-sdlc-backend

Backend source code for the car management system, implemented in Go with a
PostgreSQL database. This repository currently contains the initial database
schema and the tooling used to apply it.

## Technology Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate) with the `pgx/v5` driver
- **API style:** REST (no ORM)

## Repository Layout

```
cmd/migrate/          CLI entrypoint that applies or rolls back migrations
internal/config/       Loads database connection settings from environment variables
internal/db/           Migration runner built on golang-migrate (Migrate / Rollback)
db/migrations/         Versioned SQL migration files (one table per file, up/down pairs)
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

## Running Migrations

Connection settings are read from environment variables (see
`internal/config/config.go` for defaults):

| Variable | Default |
| --- | --- |
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_USER` | `postgres` |
| `DB_PASSWORD` | `postgres` |
| `DB_NAME` | `ai_sdlc` |
| `DB_SSLMODE` | `disable` |

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
