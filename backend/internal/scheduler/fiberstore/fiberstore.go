// Package fiberstore adapts a fiber.Storage (e.g. the Redis-backed one
// shared across this app) into a scheduler.StateStore, so Scheduler jobs
// can persist their next-run time and survive process restarts.
//
// It lives outside the scheduler package on purpose: scheduler itself has
// no dependency on Fiber or any specific storage backend, and this adapter
// is what wires it to whatever fiber.Storage the app already has.
package fiberstore

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

// defaultPrefix namespaces every key this Store writes, so it doesn't
// collide with sessions, rate limiting, or anything else sharing the same
// underlying storage.
const defaultPrefix = "scheduler:next-run:"

// Store implements scheduler.StateStore on top of a fiber.Storage.
type Store struct {
	storage fiber.Storage
	prefix  string
	ttl     time.Duration
}

// Option configures a Store built with New.
type Option func(*Store)

// WithPrefix overrides the default key prefix ("scheduler:next-run:").
func WithPrefix(prefix string) Option {
	return func(s *Store) { s.prefix = prefix }
}

// WithTTL sets an expiration on every persisted key. Defaults to 0 (no
// expiration): a job's next-run time is overwritten on every run, not
// read once and discarded, so it has no natural expiry — set this only if
// you specifically want stale entries (e.g. from a job that was removed
// from the codebase) to eventually fall out of storage on their own.
func WithTTL(ttl time.Duration) Option {
	return func(s *Store) { s.ttl = ttl }
}

// New builds a Store backed by storage.
func New(storage fiber.Storage, opts ...Option) *Store {
	s := &Store{storage: storage, prefix: defaultPrefix}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Store) key(jobName string) string {
	return s.prefix + jobName
}

// LoadNextRun implements scheduler.StateStore.
func (s *Store) LoadNextRun(ctx context.Context, jobName string) (time.Time, bool, error) {
	raw, err := s.storage.GetWithContext(ctx, s.key(jobName))
	if err != nil {
		return time.Time{}, false, fmt.Errorf("fiberstore: get %q: %w", jobName, err)
	}

	if raw == nil {
		return time.Time{}, false, nil
	}

	var next time.Time
	if err := next.UnmarshalText(raw); err != nil {
		return time.Time{}, false, fmt.Errorf("fiberstore: decode %q: %w", jobName, err)
	}

	return next, true, nil
}

// SaveNextRun implements scheduler.StateStore.
func (s *Store) SaveNextRun(ctx context.Context, jobName string, next time.Time) error {
	raw, err := next.MarshalText()
	if err != nil {
		return fmt.Errorf("fiberstore: encode %q: %w", jobName, err)
	}

	if err := s.storage.SetWithContext(ctx, s.key(jobName), raw, s.ttl); err != nil {
		return fmt.Errorf("fiberstore: set %q: %w", jobName, err)
	}

	return nil
}
