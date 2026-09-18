package vehicle

import (
	"context"
	"strings"
	"time"
)

// Service implements the Create Vehicle business logic described in the
// Vehicle Onboarding TRD: validating the request, checking category
// existence/activity, and persisting the new vehicle record.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService builds a Service backed by the given Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// Create validates req, verifies that category_id refers to an existing and
// active vehicle category, and persists the new vehicle record. actor is the
// identifier of the authenticated user performing the request, recorded in
// the audit created_by/updated_by columns.
func (s *Service) Create(ctx context.Context, req CreateRequest, actor string) (*Vehicle, error) {
	if verr := ValidateCreateRequest(req, s.now()); verr != nil {
		return nil, verr
	}

	found, active, err := s.repo.CategoryStatus(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrCategoryNotFound
	}
	if !active {
		return nil, ErrCategoryInactive
	}

	v := &Vehicle{
		VIN:            strings.TrimSpace(req.VIN),
		PlateNumber:    strings.TrimSpace(req.PlateNumber),
		Make:           strings.TrimSpace(req.Make),
		Model:          strings.TrimSpace(req.Model),
		Year:           req.Year,
		CategoryID:     req.CategoryID,
		Mileage:        req.Mileage,
		FuelLevel:      req.FuelLevel,
		Location:       strings.TrimSpace(req.Location),
		ConditionNotes: req.ConditionNotes,
		CreatedBy:      actor,
		UpdatedBy:      actor,
	}

	if err := s.repo.CreateVehicle(ctx, v); err != nil {
		return nil, err
	}

	return v, nil
}
