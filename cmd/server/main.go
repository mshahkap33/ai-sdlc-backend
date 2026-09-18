// Command server runs the ai-sdlc-backend REST API, currently exposing the
// Vehicle Onboarding "Create Vehicle" endpoint.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mshahkap33/ai-sdlc-backend/internal/auth"
	"github.com/mshahkap33/ai-sdlc-backend/internal/config"
	"github.com/mshahkap33/ai-sdlc-backend/internal/vehicle"
)

func main() {
	dbCfg := config.LoadDatabaseConfig()
	authCfg, err := config.LoadAuthConfig()
	if err != nil {
		log.Fatalf("loading auth configuration: %v", err)
	}
	serverCfg := config.LoadServerConfig()

	if authCfg.JWTPublicKeyPEM == "" {
		log.Fatal("AUTH_JWT_PUBLIC_KEY (or AUTH_JWT_PUBLIC_KEY_PATH) must be set to verify API bearer tokens")
	}
	publicKey, err := auth.ParseRSAPublicKeyFromPEM(authCfg.JWTPublicKeyPEM)
	if err != nil {
		log.Fatalf("parsing AUTH_JWT_PUBLIC_KEY: %v", err)
	}
	verifier := auth.NewVerifier(publicKey, authCfg.JWTIssuer, authCfg.JWTAudience)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbCfg.DSN())
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	repo := vehicle.NewPostgresRepository(pool)
	service := vehicle.NewService(repo)
	handler := vehicle.NewHandler(service)

	mux := http.NewServeMux()
	handler.Register(mux)

	srv := &http.Server{
		Addr:         serverCfg.Addr,
		Handler:      auth.Middleware(verifier)(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("listening on %s", serverCfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
