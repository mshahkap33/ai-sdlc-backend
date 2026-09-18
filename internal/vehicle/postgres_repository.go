package vehicle

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uniqueViolationCode is the PostgreSQL SQLSTATE for a unique constraint
// violation.
const uniqueViolationCode = "23505"

// PostgresRepository is a pgx-backed implementation of Repository.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository builds a PostgresRepository backed by the given
// connection pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CategoryStatus implements Repository.
func (r *PostgresRepository) CategoryStatus(ctx context.Context, categoryID string) (bool, bool, error) {
	var active bool
	err := r.pool.QueryRow(ctx,
		`SELECT active FROM vehicle_categories WHERE id = $1 AND deleted = false`,
		categoryID,
	).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("querying vehicle category: %w", err)
	}
	return true, active, nil
}

// CreateVehicle implements Repository.
func (r *PostgresRepository) CreateVehicle(ctx context.Context, v *Vehicle) error {
	const query = `
		INSERT INTO vehicles (
			vin, plate_number, make, model, year, category_id,
			mileage, fuel_level, location, condition_notes,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		v.VIN, v.PlateNumber, v.Make, v.Model, v.Year, v.CategoryID,
		v.Mileage, v.FuelLevel, v.Location, v.ConditionNotes,
		v.CreatedBy, v.UpdatedBy,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			switch pgErr.ConstraintName {
			case "uq_vehicles_vin":
				return ErrDuplicateVIN
			case "uq_vehicles_plate_number":
				return ErrDuplicatePlateNumber
			}
		}
		return fmt.Errorf("inserting vehicle: %w", err)
	}

	return nil
}
