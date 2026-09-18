package vehiclestatus

import "time"

// VehicleStatus is the current lifecycle status of a vehicle.
type VehicleStatus struct {
	VehicleID   string
	CategoryID  string
	Status      Status
	StatusSince time.Time
}

// Trigger is a configured mapping between an event type and the status it
// produces, as stored in the vehicle_status_triggers table.
type Trigger struct {
	EventType       EventType
	ResultingStatus Status
	RequiresReason  bool
}

// HistoryEntry is an append-only record of a single status transition,
// persisted to vehicle_status_history.
type HistoryEntry struct {
	VehicleID      string
	PreviousStatus Status
	NewStatus      Status
	TriggerEvent   string
	Reason         string
	EffectiveAt    time.Time
	Actor          string
}
