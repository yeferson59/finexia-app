package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/env"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

func main() {
	cmd := flag.String("cmd", "up", "Migration command: up | down | version")
	steps := flag.Int("steps", 1, "Number of steps for 'down' (default 1)")

	flag.Parse()

	log := logger.New(logger.Config{
		Level:       logger.LevelInfo,
		Environment: "production",
	})
	ctx := context.Background()

	if err := runMigration(ctx, log, *cmd, *steps); err != nil {
		log.Fatal(ctx, "migration failed", logger.Err(err))
	}
}

// runMigration opens the migrate instance, defers its close, then delegates to run.
// Keeping the defer here (no log.Fatal/os.Exit in this function) satisfies gocritic.
func runMigration(ctx context.Context, log logger.Logger, cmd string, steps int) error {
	_ = env.Load()

	c, err := config.New()
	if err != nil {
		return err
	}

	m, err := migrate.New(c.PathMigration, c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Error(ctx, "source close error", logger.Err(srcErr))
		}

		if dbErr != nil {
			log.Error(ctx, "db close error", logger.Err(dbErr))
		}
	}()

	return run(ctx, log, m, cmd, steps)
}

func run(ctx context.Context, log logger.Logger, m *migrate.Migrate, cmd string, steps int) error {
	switch cmd {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info(ctx, "No pending migrations.")

				return nil
			}

			return fmt.Errorf("migrate up: %w", err)
		}

		v, dirty, _ := m.Version()
		log.Info(ctx, "Migration up complete", logger.Int("version", int(v)), logger.Bool("dirty", dirty))

	case "down":
		if steps < 1 {
			return fmt.Errorf("-steps must be >= 1")
		}

		if err := m.Steps(-steps); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info(ctx, "Nothing to roll back.")

				return nil
			}

			return fmt.Errorf("migrate down %d steps: %w", steps, err)
		}

		v, dirty, verErr := m.Version()
		if verErr != nil && !errors.Is(verErr, migrate.ErrNilVersion) {
			log.Info(ctx, "Rolled back — no remaining version", logger.Int("step(s)", steps))
		} else {
			log.Info(ctx, "Rolled back", logger.Int("step(s)", steps), logger.Int("version", int(v)), logger.Bool("dirty", dirty))
		}

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				log.Info(ctx, "No migrations applied yet (version: nil)")

				return nil
			}

			return fmt.Errorf("version: %w", err)
		}

		log.Info(ctx, "Current version", logger.Int("version", int(v)), logger.Bool("dirty", dirty))

	default:
		return fmt.Errorf("unknown command %q — use: up | down | version", cmd)
	}

	return nil
}
