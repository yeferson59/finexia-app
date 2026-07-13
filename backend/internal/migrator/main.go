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

	"github.com/yeferson59/finexia-app/internal/platform/config"
)

func main() {
	cmd := flag.String("cmd", "up", "Migration command: up | down | version")
	steps := flag.Int("steps", 1, "Number of steps for 'down' (default 1)")
	flag.Parse()

	if err := runMigration(*cmd, *steps); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

// runMigration opens the migrate instance, defers its close, then delegates to run.
// Keeping the defer here (no log.Fatal/os.Exit in this function) satisfies gocritic.
func runMigration(cmd string, steps int) error {
	c := config.New()
	cfg := c.LoadEnvs()

	m, err := migrate.New(cfg.PathMigration, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
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

	return run(m, cmd, steps)
}

func run(m *migrate.Migrate, cmd string, steps int) error {
	switch cmd {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("No pending migrations.")
				return nil
			}
			return fmt.Errorf("migrate up: %w", err)
		}
		v, dirty, _ := m.Version()
		log.Printf("Migration up complete — version %d (dirty=%v)", v, dirty)

	case "down":
		if steps < 1 {
			return fmt.Errorf("-steps must be >= 1")
		}
		if err := m.Steps(-steps); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("Nothing to roll back.")
				return nil
			}
			return fmt.Errorf("migrate down %d steps: %w", steps, err)
		}
		v, dirty, verErr := m.Version()
		if verErr != nil && !errors.Is(verErr, migrate.ErrNilVersion) {
			log.Printf("Rolled back %d step(s) — no remaining version", steps)
		} else {
			log.Printf("Rolled back %d step(s) — now at version %d (dirty=%v)", steps, v, dirty)
		}

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("No migrations applied yet (version: nil)")
				return nil
			}
			return fmt.Errorf("version: %w", err)
		}
		fmt.Printf("Current version: %s  dirty: %s\n", strconv.Itoa(int(v)), strconv.FormatBool(dirty))

	default:
		return fmt.Errorf("unknown command %q — use: up | down | version", cmd)
	}

	return nil
}
