package vehicle

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRepository is an in-memory Repository used for service-level tests.
type fakeRepository struct {
	categoryFound  bool
	categoryActive bool
	categoryErr    error

	createErr error
	created   *Vehicle
}

func (f *fakeRepository) CategoryStatus(ctx context.Context, categoryID string) (bool, bool, error) {
	if f.categoryErr != nil {
		return false, false, f.categoryErr
	}
	return f.categoryFound, f.categoryActive, nil
}

func (f *fakeRepository) CreateVehicle(ctx context.Context, v *Vehicle) error {
	if f.createErr != nil {
		return f.createErr
	}
	v.ID = "11111111-1111-1111-1111-111111111111"
	v.CreatedAt = time.Date(2024, time.March, 1, 12, 0, 0, 0, time.UTC)
	v.UpdatedAt = v.CreatedAt
	f.created = v
	return nil
}

func newService(repo *fakeRepository) *Service {
	s := NewService(repo)
	s.now = func() time.Time { return time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC) }
	return s
}

func TestServiceCreateSuccess(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true}
	s := newService(repo)

	v, err := s.Create(context.Background(), validRequest(), "staff-1")
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if v.ID == "" {
		t.Fatalf("Create() vehicle ID is empty")
	}
	if v.CreatedBy != "staff-1" || v.UpdatedBy != "staff-1" {
		t.Fatalf("Create() CreatedBy/UpdatedBy = %q/%q, want %q", v.CreatedBy, v.UpdatedBy, "staff-1")
	}
}

func TestServiceCreateValidationError(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true}
	s := newService(repo)

	req := validRequest()
	req.VIN = ""

	_, err := s.Create(context.Background(), req, "staff-1")

	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("Create() error = %v, want *ValidationError", err)
	}
}

func TestServiceCreateCategoryNotFound(t *testing.T) {
	repo := &fakeRepository{categoryFound: false}
	s := newService(repo)

	_, err := s.Create(context.Background(), validRequest(), "staff-1")
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("Create() error = %v, want ErrCategoryNotFound", err)
	}
}

func TestServiceCreateCategoryInactive(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: false}
	s := newService(repo)

	_, err := s.Create(context.Background(), validRequest(), "staff-1")
	if !errors.Is(err, ErrCategoryInactive) {
		t.Fatalf("Create() error = %v, want ErrCategoryInactive", err)
	}
}

func TestServiceCreateDuplicateVIN(t *testing.T) {
	repo := &fakeRepository{categoryFound: true, categoryActive: true, createErr: ErrDuplicateVIN}
	s := newService(repo)

	_, err := s.Create(context.Background(), validRequest(), "staff-1")
	if !errors.Is(err, ErrDuplicateVIN) {
		t.Fatalf("Create() error = %v, want ErrDuplicateVIN", err)
	}
}
