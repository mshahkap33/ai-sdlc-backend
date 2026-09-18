// Command migrate applies or rolls back the car management database schema
// migrations against a PostgreSQL database.
package main

import (
	"flag"
	"log"

	"github.com/mshahkap33/ai-sdlc-backend/internal/config"
	"github.com/mshahkap33/ai-sdlc-backend/internal/db"
)

func main() {
	var (
		direction      string
		steps          int
		migrationsPath string
	)

	flag.StringVar(&direction, "direction", "up", "migration direction: up or down")
	flag.IntVar(&steps, "steps", 1, "number of migrations to roll back when direction=down")
	flag.StringVar(&migrationsPath, "path", db.MigrationsPath, "path to the migration files")
	flag.Parse()

	cfg := config.LoadDatabaseConfig()
	dsn := cfg.DSN()

	var err error
	switch direction {
	case "up":
		err = db.Migrate(dsn, migrationsPath)
	case "down":
		err = db.Rollback(dsn, migrationsPath, steps)
	default:
		log.Fatalf("unknown direction %q: must be \"up\" or \"down\"", direction)
	}

	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Printf("migration %s completed successfully", direction)
}
