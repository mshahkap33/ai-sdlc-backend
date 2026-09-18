package vehicle

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// postgres unique constraint names, as defined in
// db/migrations/000003_create_vehicles_table.up.sql.
const (
	uniqueVINConstraint   = "uq_vehicles_vin"
	uniquePlateConstraint = "uq_vehicles_plate_number"

	uniqueViolationCode = "23505"
)

// PostgresRepository is a Repository implementation backed by PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgresRepository that executes queries
// against pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateVehicle inserts a new vehicle row, verifying that its category is
// an existing, active category and that the VIN/plate number are not
// already in use.
func (r *PostgresRepository) CreateVehicle(ctx context.Context, v Vehicle) (Vehicle, error) {
	const query = `
		INSERT INTO vehicles (
			vin, plate_number, make, model, year, category_id,
			mileage, fuel_level, location, condition_notes,
			created_by, updated_by
		)
		SELECT
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11
		WHERE EXISTS (
			SELECT 1 FROM vehicle_categories
			WHERE id = $6 AND active = true AND deleted = false
		)
		RETURNING id, created_at`

	var conditionNotes any
	if v.ConditionNotes != "" {
		conditionNotes = v.ConditionNotes
	}

	row := r.pool.QueryRow(ctx, query,
		v.VIN, v.PlateNumber, v.Make, v.Model, v.Year, v.CategoryID,
		v.Mileage, v.FuelLevel, v.Location, conditionNotes, v.CreatedBy,
	)

	if err := row.Scan(&v.ID, &v.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Vehicle{}, ErrCategoryNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			switch pgErr.ConstraintName {
			case uniqueVINConstraint:
				return Vehicle{}, ErrDuplicateVIN
			case uniquePlateConstraint:
				return Vehicle{}, ErrDuplicatePlateNumber
			}
		}

		return Vehicle{}, err
	}

	return v, nil
}
