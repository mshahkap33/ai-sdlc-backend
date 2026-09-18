package vehiclestatus

import (
	"context"
	"fmt"
	"time"
)

// Service implements the business logic for transitioning a vehicle's
// lifecycle status, both via triggering events and manual overrides, as
// defined by the vehicle status and availability TRD.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService creates a Service backed by the given Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// ApplyEventInput is the input to ApplyEvent.
type ApplyEventInput struct {
	VehicleID    string
	EventType    EventType
	EffectiveAt  time.Time
	Reason       string
	SourceSystem string
	Actor        string
}

// TransitionResult is the outcome of a successful status transition.
type TransitionResult struct {
	VehicleID   string
	Status      Status
	StatusSince time.Time
}

// ApplyEvent submits a triggering event for a vehicle (e.g. "booking_confirmed")
// and, if the transition is allowed from the vehicle's current status,
// updates its status and records the change in the audit trail.
func (s *Service) ApplyEvent(ctx context.Context, in ApplyEventInput) (TransitionResult, error) {
	current, err := s.repo.GetVehicleStatus(ctx, in.VehicleID)
	if err != nil {
		return TransitionResult{}, err
	}

	trigger, err := s.repo.GetTrigger(ctx, in.EventType)
	if err != nil {
		return TransitionResult{}, err
	}

	if trigger.RequiresReason && in.Reason == "" {
		return TransitionResult{}, ErrReasonRequired
	}

	next, allowed := NextStatus(current.Status, in.EventType)
	if !allowed || next != trigger.ResultingStatus {
		return TransitionResult{}, ErrTransitionNotAllowed
	}

	effectiveAt := in.EffectiveAt
	if effectiveAt.IsZero() {
		effectiveAt = s.now()
	}

	entry := HistoryEntry{
		VehicleID:      in.VehicleID,
		PreviousStatus: current.Status,
		NewStatus:      next,
		TriggerEvent:   string(in.EventType),
		Reason:         in.Reason,
		EffectiveAt:    effectiveAt,
		Actor:          actorOrSource(in.Actor, in.SourceSystem),
	}

	if err := s.repo.ApplyTransition(ctx, in.VehicleID, next, effectiveAt, entry); err != nil {
		return TransitionResult{}, err
	}

	return TransitionResult{VehicleID: in.VehicleID, Status: next, StatusSince: effectiveAt}, nil
}

// OverrideInput is the input to Override.
type OverrideInput struct {
	VehicleID string
	NewStatus Status
	Reason    string
	Actor     string
}

// Override manually sets a vehicle's status, requiring a non-empty reason
// and a transition allowed by the state machine from the vehicle's current
// status.
func (s *Service) Override(ctx context.Context, in OverrideInput) (TransitionResult, error) {
	if in.Reason == "" {
		return TransitionResult{}, ErrReasonRequired
	}
	if !IsValidStatus(in.NewStatus) {
		return TransitionResult{}, ErrInvalidStatus
	}

	current, err := s.repo.GetVehicleStatus(ctx, in.VehicleID)
	if err != nil {
		return TransitionResult{}, err
	}

	if !IsTransitionAllowed(current.Status, in.NewStatus) {
		return TransitionResult{}, ErrTransitionNotAllowed
	}

	effectiveAt := s.now()

	entry := HistoryEntry{
		VehicleID:      in.VehicleID,
		PreviousStatus: current.Status,
		NewStatus:      in.NewStatus,
		TriggerEvent:   "manual_override",
		Reason:         in.Reason,
		EffectiveAt:    effectiveAt,
		Actor:          in.Actor,
	}

	if err := s.repo.ApplyTransition(ctx, in.VehicleID, in.NewStatus, effectiveAt, entry); err != nil {
		return TransitionResult{}, err
	}

	return TransitionResult{VehicleID: in.VehicleID, Status: in.NewStatus, StatusSince: effectiveAt}, nil
}

func actorOrSource(actor, sourceSystem string) string {
	if actor != "" {
		return actor
	}
	return fmt.Sprintf("system:%s", sourceSystem)
}
