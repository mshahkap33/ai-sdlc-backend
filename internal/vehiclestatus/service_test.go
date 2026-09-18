package vehiclestatus

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceApplyEventSuccess(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusAvailable, StatusSince: time.Now()}
	repo.triggers["booking_confirmed"] = Trigger{EventType: "booking_confirmed", ResultingStatus: StatusReserved}

	svc := NewService(repo)

	result, err := svc.ApplyEvent(context.Background(), ApplyEventInput{
		VehicleID:    "v1",
		EventType:    "booking_confirmed",
		SourceSystem: "booking",
	})
	if err != nil {
		t.Fatalf("ApplyEvent() error = %v", err)
	}
	if result.Status != StatusReserved {
		t.Errorf("result.Status = %q, want %q", result.Status, StatusReserved)
	}
	if len(repo.history) != 1 {
		t.Fatalf("len(repo.history) = %d, want 1", len(repo.history))
	}
	if repo.history[0].Actor != "system:booking" {
		t.Errorf("history actor = %q, want %q", repo.history[0].Actor, "system:booking")
	}
	if repo.vehicles["v1"].Status != StatusReserved {
		t.Errorf("vehicle status = %q, want %q", repo.vehicles["v1"].Status, StatusReserved)
	}
}

func TestServiceApplyEventVehicleNotFound(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)

	_, err := svc.ApplyEvent(context.Background(), ApplyEventInput{VehicleID: "missing", EventType: "booking_confirmed"})
	if !errors.Is(err, ErrVehicleNotFound) {
		t.Fatalf("ApplyEvent() error = %v, want ErrVehicleNotFound", err)
	}
}

func TestServiceApplyEventUnknownEventType(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusAvailable}
	svc := NewService(repo)

	_, err := svc.ApplyEvent(context.Background(), ApplyEventInput{VehicleID: "v1", EventType: "not_configured"})
	if !errors.Is(err, ErrUnknownEventType) {
		t.Fatalf("ApplyEvent() error = %v, want ErrUnknownEventType", err)
	}
}

func TestServiceApplyEventTransitionNotAllowed(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusRented}
	repo.triggers["booking_confirmed"] = Trigger{EventType: "booking_confirmed", ResultingStatus: StatusReserved}
	svc := NewService(repo)

	// booking_confirmed is not a valid transition from "rented".
	_, err := svc.ApplyEvent(context.Background(), ApplyEventInput{VehicleID: "v1", EventType: "booking_confirmed"})
	if !errors.Is(err, ErrTransitionNotAllowed) {
		t.Fatalf("ApplyEvent() error = %v, want ErrTransitionNotAllowed", err)
	}
}

func TestServiceApplyEventRequiresReason(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusAvailable}
	repo.triggers["damage_reported"] = Trigger{EventType: "damage_reported", ResultingStatus: StatusDamaged, RequiresReason: true}
	svc := NewService(repo)

	_, err := svc.ApplyEvent(context.Background(), ApplyEventInput{VehicleID: "v1", EventType: "damage_reported"})
	if !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("ApplyEvent() error = %v, want ErrReasonRequired", err)
	}

	_, err = svc.ApplyEvent(context.Background(), ApplyEventInput{VehicleID: "v1", EventType: "damage_reported", Reason: "collision"})
	if err != nil {
		t.Fatalf("ApplyEvent() with reason error = %v", err)
	}
}

func TestServiceOverrideSuccess(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusAvailable}
	svc := NewService(repo)

	result, err := svc.Override(context.Background(), OverrideInput{
		VehicleID: "v1",
		NewStatus: StatusMaintenance,
		Reason:    "scheduled service",
		Actor:     "staff-1",
	})
	if err != nil {
		t.Fatalf("Override() error = %v", err)
	}
	if result.Status != StatusMaintenance {
		t.Errorf("result.Status = %q, want %q", result.Status, StatusMaintenance)
	}
	if repo.history[0].TriggerEvent != "manual_override" {
		t.Errorf("history trigger event = %q, want %q", repo.history[0].TriggerEvent, "manual_override")
	}
}

func TestServiceOverrideRequiresReason(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusAvailable}
	svc := NewService(repo)

	_, err := svc.Override(context.Background(), OverrideInput{VehicleID: "v1", NewStatus: StatusMaintenance})
	if !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("Override() error = %v, want ErrReasonRequired", err)
	}
}

func TestServiceOverrideInvalidStatus(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusAvailable}
	svc := NewService(repo)

	_, err := svc.Override(context.Background(), OverrideInput{VehicleID: "v1", NewStatus: "bogus", Reason: "x"})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("Override() error = %v, want ErrInvalidStatus", err)
	}
}

func TestServiceOverrideTransitionNotAllowed(t *testing.T) {
	repo := newFakeRepository()
	repo.vehicles["v1"] = VehicleStatus{VehicleID: "v1", Status: StatusRetired}
	svc := NewService(repo)

	_, err := svc.Override(context.Background(), OverrideInput{VehicleID: "v1", NewStatus: StatusAvailable, Reason: "reactivate"})
	if !errors.Is(err, ErrTransitionNotAllowed) {
		t.Fatalf("Override() error = %v, want ErrTransitionNotAllowed", err)
	}
}
