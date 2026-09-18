package maintenanceschedule

import (
	"context"
	"testing"
)

type fakeRepository struct {
	createFunc func(ctx context.Context, s Schedule) (Schedule, error)
	called     bool
	received   Schedule
}

func (f *fakeRepository) CreateSchedule(ctx context.Context, s Schedule) (Schedule, error) {
	f.called = true
	f.received = s
	if f.createFunc != nil {
		return f.createFunc(ctx, s)
	}
	return s, nil
}

func TestServiceCreateScheduleValid(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, s Schedule) (Schedule, error) {
			s.ID = "generated-id"
			s.Active = true
			return s, nil
		},
	}
	s := NewService(repo)

	created, err := s.CreateSchedule(context.Background(), validDateIntervalRequest(), "staff-1")
	if err != nil {
		t.Fatalf("CreateSchedule() error = %v", err)
	}
	if !repo.called {
		t.Fatal("expected repository.CreateSchedule to be called")
	}
	if created.ID != "generated-id" {
		t.Errorf("created.ID = %q, want %q", created.ID, "generated-id")
	}
	if repo.received.CreatedBy != "staff-1" {
		t.Errorf("received.CreatedBy = %q, want %q", repo.received.CreatedBy, "staff-1")
	}
}

func TestServiceCreateScheduleInvalidRequestDoesNotCallRepository(t *testing.T) {
	repo := &fakeRepository{}
	s := NewService(repo)

	req := validDateIntervalRequest()
	req.IntervalDays = nil

	_, err := s.CreateSchedule(context.Background(), req, "staff-1")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if repo.called {
		t.Fatal("expected repository.CreateSchedule not to be called for invalid request")
	}
}

func TestServiceCreateSchedulePropagatesRepositoryError(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, s Schedule) (Schedule, error) {
			return Schedule{}, ErrVehicleCategoryNotFound
		},
	}
	s := NewService(repo)

	_, err := s.CreateSchedule(context.Background(), validDateIntervalRequest(), "staff-1")
	if err != ErrVehicleCategoryNotFound {
		t.Fatalf("CreateSchedule() error = %v, want %v", err, ErrVehicleCategoryNotFound)
	}
}
