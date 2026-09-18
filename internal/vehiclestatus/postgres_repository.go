package vehiclestatus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository is a pgx-backed implementation of Repository for
// PostgreSQL, using the tables defined by the vehicle status and
// availability migrations (vehicles, vehicle_status_triggers,
// vehicle_status_history).
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgresRepository that executes queries
// against pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// GetVehicleStatus implements Repository.
func (r *PostgresRepository) GetVehicleStatus(ctx context.Context, vehicleID string) (VehicleStatus, error) {
	const query = `
		SELECT id, category_id, status, status_since
		FROM vehicles
		WHERE id = $1 AND deleted = false
	`

	var vs VehicleStatus
	err := r.pool.QueryRow(ctx, query, vehicleID).Scan(&vs.VehicleID, &vs.CategoryID, &vs.Status, &vs.StatusSince)
	if errors.Is(err, pgx.ErrNoRows) {
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
	err := r.pool.QueryRow(ctx, query, string(eventType)).Scan(&t.EventType, &t.ResultingStatus, &t.RequiresReason)
	if errors.Is(err, pgx.ErrNoRows) {
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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const updateVehicle = `
		UPDATE vehicles
		SET status = $1, status_since = $2, updated_at = now(), updated_by = $3
		WHERE id = $4 AND deleted = false
	`
	tag, err := tx.Exec(ctx, updateVehicle, string(newStatus), statusSince, entry.Actor, vehicleID)
	if err != nil {
		return fmt.Errorf("updating vehicle status: %w", err)
	}
	if tag.RowsAffected() == 0 {
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
	if _, err := tx.Exec(ctx, insertHistory,
		vehicleID, previous, string(entry.NewStatus), entry.TriggerEvent, reason, entry.EffectiveAt, entry.Actor,
	); err != nil {
		return fmt.Errorf("inserting status history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
