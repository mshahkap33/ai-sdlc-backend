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
	"github.com/mshahkap33/ai-sdlc-backend/internal/vehicle"
)

// serviceStaffRole is the role required to create or update vehicle master
// data, as required by the Vehicle Onboarding TRD's security requirements.
const serviceStaffRole = "service_staff"

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

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/vehicles", verifier.Authenticate(
		auth.RequireRole(serviceStaffRole, http.HandlerFunc(vehicleHandler.CreateVehicle)),
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
