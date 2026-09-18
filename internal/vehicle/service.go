package vehicle

import (
	"context"
	"time"
)

// Service implements the Create Vehicle use case described in the Vehicle
// Onboarding TRD: validating the request and persisting the new vehicle
// record.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService creates a Service backed by repo.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// CreateVehicle validates req and, if valid, persists a new vehicle record
// created by the given actor (the authenticated staff user's subject).
func (s *Service) CreateVehicle(ctx context.Context, req CreateRequest, actor string) (Vehicle, error) {
	if errs := req.validate(s.now()); len(errs) > 0 {
		return Vehicle{}, errs
	}

	v := Vehicle{
		VIN:            req.VIN,
		PlateNumber:    req.PlateNumber,
		Make:           req.Make,
		Model:          req.Model,
		Year:           req.Year,
		CategoryID:     req.CategoryID,
		Mileage:        req.Mileage,
		FuelLevel:      req.FuelLevel,
		Location:       req.Location,
		ConditionNotes: req.ConditionNotes,
		CreatedBy:      actor,
	}

	return s.repo.CreateVehicle(ctx, v)
}
