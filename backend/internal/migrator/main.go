package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/yeferson59/finexia-app/internal/config"
)

func main() {
	cmd := flag.String("cmd", "up", "Migration command: up | down | version")
	steps := flag.Int("steps", 1, "Number of steps for 'down' (default 1)")
	flag.Parse()

	c := config.New()
	cfg := c.LoadEnvs()

	m, err := migrate.New(cfg.PathMigration, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("migrate.New: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("source close error: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("db close error: %v", dbErr)
		}
	}()

	switch *cmd {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("No pending migrations.")
				return
			}
			log.Fatalf("migrate up: %v", err)
		}
		v, dirty, _ := m.Version()
		log.Printf("Migration up complete — version %d (dirty=%v)", v, dirty)

	case "down":
		if *steps < 1 {
			log.Fatal("-steps must be >= 1")
		}
		if err := m.Steps(-(*steps)); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("Nothing to roll back.")
				return
			}
			log.Fatalf("migrate down %d steps: %v", *steps, err)
		}
		v, dirty, verErr := m.Version()
		if verErr != nil && !errors.Is(verErr, migrate.ErrNilVersion) {
			log.Printf("rolled back %d step(s) — no remaining version", *steps)
		} else {
			log.Printf("Rolled back %d step(s) — now at version %d (dirty=%v)", *steps, v, dirty)
		}

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("No migrations applied yet (version: nil)")
				return
			}
			log.Fatalf("version: %v", err)
		}
		fmt.Printf("Current version: %s  dirty: %s\n", strconv.Itoa(int(v)), strconv.FormatBool(dirty))

	default:
		log.Fatalf("Unknown command %q — use: up | down | version", *cmd)
	}
}
