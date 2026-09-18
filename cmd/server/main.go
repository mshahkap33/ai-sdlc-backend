// Command server starts the car management REST API HTTP server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
	"github.com/mshahkap33/ai-sdlc-backend/internal/config"
	"github.com/mshahkap33/ai-sdlc-backend/internal/maintenanceschedule"
	"github.com/mshahkap33/ai-sdlc-backend/internal/vehicle"
	"github.com/mshahkap33/ai-sdlc-backend/internal/vehiclestatus"
)

// Role names recognized by the JWT `roles` claim, per the Vehicle
// Onboarding and Vehicle Status and Availability TRDs' security
// requirements.
const (
	roleServiceStaff      = "service_staff"
	roleOperationsManager = "operations_manager"
	roleSystemService     = "system_service"
)

func main() {
	dbCfg := config.LoadDatabaseConfig()

	authCfg, err := config.LoadAuthConfig()
	if err != nil {
		log.Fatalf("loading auth config: %v", err)
	}

	publicKey, err := authCfg.PublicKey()
	if err != nil {
		log.Fatalf("loading JWT public key: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbCfg.DSN())
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	verifier := auth.NewVerifier(publicKey, authCfg.Issuer, authCfg.Audience)

	vehicleRepo := vehicle.NewPostgresRepository(pool)
	vehicleService := vehicle.NewService(vehicleRepo)
	vehicleHandler := vehicle.NewHandler(vehicleService)

	statusRepo := vehiclestatus.NewPostgresRepository(pool)
	statusService := vehiclestatus.NewService(statusRepo)
	statusHandler := vehiclestatus.NewHandler(statusService)

	scheduleRepo := maintenanceschedule.NewPostgresRepository(pool)
	scheduleService := maintenanceschedule.NewService(scheduleRepo)
	scheduleHandler := maintenanceschedule.NewHandler(scheduleService)

	// vehiclestatus.Handler registers its routes on its own mux so that the
	// status-events and status (override) endpoints can be wrapped with
	// different role requirements below.
	statusRoutes := http.NewServeMux()
	statusHandler.Register(statusRoutes)

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/vehicles", verifier.Authenticate(
		auth.RequireRole(roleServiceStaff, http.HandlerFunc(vehicleHandler.CreateVehicle)),
	))

	// status-events is a system-driven transition restricted to trusted
	// internal service accounts; status (manual override) requires service
	// staff or operations manager privileges.
	mux.Handle("POST /api/v1/vehicles/{vehicleId}/status-events", verifier.Authenticate(
		auth.RequireRole(roleSystemService, statusRoutes),
	))
	mux.Handle("PATCH /api/v1/vehicles/{vehicleId}/status", verifier.Authenticate(
		auth.RequireAnyRole([]string{roleServiceStaff, roleOperationsManager}, statusRoutes),
	))

	mux.Handle("POST /api/v1/maintenance-schedules", verifier.Authenticate(
		auth.RequireAnyRole([]string{roleServiceStaff, roleOperationsManager}, http.HandlerFunc(scheduleHandler.CreateSchedule)),
	))

	addr := getEnv("HTTP_ADDR", ":8080")
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
