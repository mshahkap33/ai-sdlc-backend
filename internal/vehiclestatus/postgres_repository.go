package vehiclestatus

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PostgresRepository is a database/sql-backed implementation of Repository
// for PostgreSQL, using the tables defined by the vehicle status and
// availability migrations (vehicles, vehicle_status_triggers,
// vehicle_status_history).
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository wraps an existing *sql.DB connection pool.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetVehicleStatus implements Repository.
func (r *PostgresRepository) GetVehicleStatus(ctx context.Context, vehicleID string) (VehicleStatus, error) {
	const query = `
		SELECT id, category_id, status, status_since
		FROM vehicles
		WHERE id = $1 AND deleted = false
	`

	var vs VehicleStatus
	err := r.db.QueryRowContext(ctx, query, vehicleID).Scan(&vs.VehicleID, &vs.CategoryID, &vs.Status, &vs.StatusSince)
	if errors.Is(err, sql.ErrNoRows) {
		return VehicleStatus{}, ErrVehicleNotFound
	}
	if err != nil {
		return VehicleStatus{}, fmt.Errorf("querying vehicle status: %w", err)
	}

	return vs, nil
}

// GetTrigger implements Repository.
func (r *PostgresRepository) GetTrigger(ctx context.Context, eventType EventType) (Trigger, error) {
	const query = `
		SELECT event_type, resulting_status, requires_reason
		FROM vehicle_status_triggers
		WHERE event_type = $1 AND deleted = false
	`

	var t Trigger
	err := r.db.QueryRowContext(ctx, query, string(eventType)).Scan(&t.EventType, &t.ResultingStatus, &t.RequiresReason)
	if errors.Is(err, sql.ErrNoRows) {
		return Trigger{}, ErrUnknownEventType
	}
	if err != nil {
		return Trigger{}, fmt.Errorf("querying vehicle status trigger: %w", err)
	}

	return t, nil
}

// ApplyTransition implements Repository. It updates the vehicle's status and
// appends the audit trail entry within a single database transaction.
func (r *PostgresRepository) ApplyTransition(ctx context.Context, vehicleID string, newStatus Status, statusSince time.Time, entry HistoryEntry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const updateVehicle = `
		UPDATE vehicles
		SET status = $1, status_since = $2, updated_at = now(), updated_by = $3
		WHERE id = $4 AND deleted = false
	`
	res, err := tx.ExecContext(ctx, updateVehicle, string(newStatus), statusSince, entry.Actor, vehicleID)
	if err != nil {
		return fmt.Errorf("updating vehicle status: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("checking updated rows: %w", err)
	} else if rows == 0 {
		return ErrVehicleNotFound
	}

	const insertHistory = `
		INSERT INTO vehicle_status_history
			(vehicle_id, previous_status, new_status, trigger_event, reason, effective_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`
	var previous interface{}
	if entry.PreviousStatus != "" {
		previous = string(entry.PreviousStatus)
	}
	var reason interface{}
	if entry.Reason != "" {
		reason = entry.Reason
	}
	if _, err := tx.ExecContext(ctx, insertHistory,
		vehicleID, previous, string(entry.NewStatus), entry.TriggerEvent, reason, entry.EffectiveAt, entry.Actor,
	); err != nil {
		return fmt.Errorf("inserting status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
