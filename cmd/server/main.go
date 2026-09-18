// Command server runs the REST API for the car management service,
// currently exposing the vehicle lifecycle status transition endpoints.
package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
	"github.com/mshahkap33/ai-sdlc-backend/internal/config"
	"github.com/mshahkap33/ai-sdlc-backend/internal/vehiclestatus"
)

// Role names recognized by the JWT `roles` claim, per the vehicle status
// and availability TRD's security requirements.
const (
	roleServiceStaff      = "service_staff"
	roleOperationsManager = "operations_manager"
	roleSystemService     = "system_service"
)

func main() {
	dbCfg := config.LoadDatabaseConfig()
	serverCfg := config.LoadServerConfig()

	if serverCfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set to run the server")
	}

	db, err := sql.Open("pgx", dbCfg.DSN())
	if err != nil {
		log.Fatalf("opening database connection: %v", err)
	}
	defer db.Close()

	verifier := auth.NewVerifier([]byte(serverCfg.JWTSecret))
	repo := vehiclestatus.NewPostgresRepository(db)
	service := vehiclestatus.NewService(repo)
	handler := vehiclestatus.NewHandler(service)

	statusRoutes := http.NewServeMux()
	handler.Register(statusRoutes)

	mux := http.NewServeMux()

	// POST /status-events is restricted to trusted internal service
	// accounts; PATCH /status (manual override) requires service staff or
	// operations manager privileges.
	mux.Handle("POST /api/v1/vehicles/{vehicleId}/status-events",
		verifier.RequireRoles(roleSystemService)(statusRoutes))
	mux.Handle("PATCH /api/v1/vehicles/{vehicleId}/status",
		verifier.RequireRoles(roleServiceStaff, roleOperationsManager)(statusRoutes))

	log.Printf("listening on %s", serverCfg.Addr)
	if err := http.ListenAndServe(serverCfg.Addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
