package vehicle

import (
	"context"
	"testing"
	"time"
)

type fakeRepository struct {
	createFunc func(ctx context.Context, v Vehicle) (Vehicle, error)
	called     bool
	received   Vehicle
}

func (f *fakeRepository) CreateVehicle(ctx context.Context, v Vehicle) (Vehicle, error) {
	f.called = true
	f.received = v
	if f.createFunc != nil {
		return f.createFunc(ctx, v)
	}
	return v, nil
}

func newTestService(repo Repository) *Service {
	s := NewService(repo)
	s.now = func() time.Time { return time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC) }
	return s
}

func TestServiceCreateVehicleValid(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, v Vehicle) (Vehicle, error) {
			v.ID = "generated-id"
			v.CreatedAt = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
			return v, nil
		},
	}
	s := newTestService(repo)

	created, err := s.CreateVehicle(context.Background(), validRequest(), "staff-1")
	if err != nil {
		t.Fatalf("CreateVehicle() error = %v", err)
	}
	if !repo.called {
		t.Fatal("expected repository.CreateVehicle to be called")
	}
	if created.ID != "generated-id" {
		t.Errorf("created.ID = %q, want %q", created.ID, "generated-id")
	}
	if repo.received.CreatedBy != "staff-1" {
		t.Errorf("received.CreatedBy = %q, want %q", repo.received.CreatedBy, "staff-1")
	}
}

func TestServiceCreateVehicleInvalidRequestDoesNotCallRepository(t *testing.T) {
	repo := &fakeRepository{}
	s := newTestService(repo)

	req := validRequest()
	req.VIN = ""

	_, err := s.CreateVehicle(context.Background(), req, "staff-1")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if repo.called {
		t.Fatal("expected repository.CreateVehicle not to be called for invalid request")
	}
}

func TestServiceCreateVehiclePropagatesRepositoryError(t *testing.T) {
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, v Vehicle) (Vehicle, error) {
			return Vehicle{}, ErrDuplicateVIN
		},
	}
	s := newTestService(repo)

	_, err := s.CreateVehicle(context.Background(), validRequest(), "staff-1")
	if err != ErrDuplicateVIN {
		t.Fatalf("CreateVehicle() error = %v, want %v", err, ErrDuplicateVIN)
	}
}
