// Command migrate applies (or rolls back) database migrations without
// starting the HTTP server. Useful for CI and local scripts — the API also
// runs migrations automatically on boot, so this is mostly for explicit
// control (e.g. `pnpm db:migrate:down`).
package main

import (
	"flag"
	"log"

	"github.com/thiago/todo-simple-api/internal/config"
	"github.com/thiago/todo-simple-api/internal/database"
)

func main() {
	down := flag.Bool("down", false, "roll back the most recently applied migration")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if *down {
		if err := database.MigrateDown(cfg.DatabaseURL); err != nil {
			log.Fatal(err)
		}
		log.Println("rolled back one migration")
		return
	}

	if err := database.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations applied")
}
