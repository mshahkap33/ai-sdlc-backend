package vehiclestatus

import (
	"context"
	"time"
)

// fakeRepository is an in-memory Repository used by service and handler
// tests, avoiding any dependency on a real PostgreSQL instance.
type fakeRepository struct {
	vehicles map[string]VehicleStatus
	triggers map[EventType]Trigger
	history  []HistoryEntry

	applyErr error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		vehicles: map[string]VehicleStatus{},
		triggers: map[EventType]Trigger{},
	}
}

func (f *fakeRepository) GetVehicleStatus(_ context.Context, vehicleID string) (VehicleStatus, error) {
	vs, ok := f.vehicles[vehicleID]
	if !ok {
		return VehicleStatus{}, ErrVehicleNotFound
	}
	return vs, nil
}

func (f *fakeRepository) GetTrigger(_ context.Context, eventType EventType) (Trigger, error) {
	t, ok := f.triggers[eventType]
	if !ok {
		return Trigger{}, ErrUnknownEventType
	}
	return t, nil
}

func (f *fakeRepository) ApplyTransition(_ context.Context, vehicleID string, newStatus Status, statusSince time.Time, entry HistoryEntry) error {
	if f.applyErr != nil {
		return f.applyErr
	}
	vs, ok := f.vehicles[vehicleID]
	if !ok {
		return ErrVehicleNotFound
	}
	vs.Status = newStatus
	vs.StatusSince = statusSince
	f.vehicles[vehicleID] = vs
	f.history = append(f.history, entry)
	return nil
}
