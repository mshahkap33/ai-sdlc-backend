// Command migrate applies or rolls back the database schema migrations
// located in db/migrations against the PostgreSQL database configured via
// environment variables (see internal/config).
package main

import (
	"flag"
	"log"

	"github.com/mshahkap33/ai-sdlc-backend/internal/config"
	"github.com/mshahkap33/ai-sdlc-backend/internal/database"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: 'up' or 'down'")
	sourceURL := flag.String("source", database.MigrationsSourceURL, "golang-migrate source URL for the migration files")
	flag.Parse()

	dsn := config.LoadDatabaseConfig().DSN()

	var err error
	switch *direction {
	case "up":
		err = database.Up(*sourceURL, dsn)
	case "down":
		err = database.Down(*sourceURL, dsn)
	default:
		log.Fatalf("unknown direction %q: must be 'up' or 'down'", *direction)
	}

	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Printf("migration %q completed successfully", *direction)
}
