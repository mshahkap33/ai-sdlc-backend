package vehiclestatus

import (
	"context"
	"time"
)

// Repository is the persistence boundary used by Service. It is
// implemented by PostgresRepository for production use and can be faked in
// tests.
type Repository interface {
	// GetVehicleStatus loads the current status of a non-deleted vehicle.
	// It returns ErrVehicleNotFound if no such vehicle exists.
	GetVehicleStatus(ctx context.Context, vehicleID string) (VehicleStatus, error)

	// GetTrigger loads the configured status transition for an event type.
	// It returns ErrUnknownEventType if the event type is not configured.
	GetTrigger(ctx context.Context, eventType EventType) (Trigger, error)

	// ApplyTransition atomically updates the vehicle's current status and
	// appends the corresponding entry to the status history audit trail.
	ApplyTransition(ctx context.Context, vehicleID string, newStatus Status, statusSince time.Time, entry HistoryEntry) error
}
