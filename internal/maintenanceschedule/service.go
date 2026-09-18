package maintenanceschedule

import "context"

// Service implements the Create Maintenance Schedule use case described in
// the Maintenance and Service Scheduling TRD: validating the request and
// persisting the new schedule rule.
type Service struct {
	repo Repository
}

// NewService creates a Service backed by repo.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateSchedule validates req and, if valid, persists a new maintenance
// schedule rule created by the given actor (the authenticated staff user's
// subject).
func (s *Service) CreateSchedule(ctx context.Context, req CreateRequest, actor string) (Schedule, error) {
	if errs := req.validate(); len(errs) > 0 {
		return Schedule{}, errs
	}

	sched := Schedule{
		VehicleCategoryID:     req.VehicleCategoryID,
		VehicleID:             req.VehicleID,
		TriggerType:           TriggerType(req.TriggerType),
		IntervalDays:          req.IntervalDays,
		IntervalMileage:       req.IntervalMileage,
		ManufacturerReference: req.ManufacturerReference,
		TelematicsCondition:   req.TelematicsCondition,
		CreatedBy:             actor,
	}

	return s.repo.CreateSchedule(ctx, sched)
}
