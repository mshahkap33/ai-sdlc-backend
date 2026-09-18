package maintenanceschedule

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository is a Repository implementation backed by PostgreSQL,
// using the maintenance_schedules table.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgresRepository that executes queries
// against pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateSchedule inserts a new maintenance schedule row, verifying that the
// referenced vehicle category or vehicle exists and is not deleted.
func (r *PostgresRepository) CreateSchedule(ctx context.Context, s Schedule) (Schedule, error) {
	if s.VehicleCategoryID != "" {
		exists, err := r.exists(ctx, "vehicle_categories", s.VehicleCategoryID)
		if err != nil {
			return Schedule{}, err
		}
		if !exists {
			return Schedule{}, ErrVehicleCategoryNotFound
		}
	}

	if s.VehicleID != "" {
		exists, err := r.exists(ctx, "vehicles", s.VehicleID)
		if err != nil {
			return Schedule{}, err
		}
		if !exists {
			return Schedule{}, ErrVehicleNotFound
		}
	}

	const query = `
		INSERT INTO maintenance_schedules (
			vehicle_category_id, vehicle_id, trigger_type,
			interval_days, interval_mileage, manufacturer_reference, telematics_condition,
			created_by, updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, active`

	var categoryID, vehicleID any
	if s.VehicleCategoryID != "" {
		categoryID = s.VehicleCategoryID
	}
	if s.VehicleID != "" {
		vehicleID = s.VehicleID
	}
	var manufacturerReference, telematicsCondition any
	if s.ManufacturerReference != "" {
		manufacturerReference = s.ManufacturerReference
	}
	if s.TelematicsCondition != "" {
		telematicsCondition = s.TelematicsCondition
	}

	row := r.pool.QueryRow(ctx, query,
		categoryID, vehicleID, string(s.TriggerType),
		s.IntervalDays, s.IntervalMileage, manufacturerReference, telematicsCondition,
		s.CreatedBy,
	)

	if err := row.Scan(&s.ID, &s.Active); err != nil {
		return Schedule{}, fmt.Errorf("inserting maintenance schedule: %w", err)
	}

	return s, nil
}

// exists reports whether a non-deleted row with the given id exists in
// table. table is a fixed, code-controlled value (never user input).
func (r *PostgresRepository) exists(ctx context.Context, table, id string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE id = $1 AND deleted = false`, table)

	var one int
	err := r.pool.QueryRow(ctx, query, id).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking existence in %s: %w", table, err)
	}
	return true, nil
}
