package scheduler

import (
	"context"
	"time"
)

// StateStore persists each job's next scheduled run time, so a Scheduler
// can resume its cadence across process restarts instead of recomputing
// every Schedule from time.Now() at Start(). It is entirely optional: a
// Scheduler with no Store configured (the default) behaves exactly as
// before — every job's first run after Start() is computed fresh, so an
// Every{Interval: 3 * time.Hour} job restarts its 3h count from scratch on
// every deploy/restart. With a Store, that job instead resumes counting
// from whenever it was last due.
//
// The scheduler package intentionally has no opinion on where the state
// lives — a StateStore can be backed by Postgres, Redis, a local file, or
// anything else; implement this interface against whatever storage the
// application already has.
//
// Implementations must be safe for concurrent use: each registered job
// calls Load/Save from its own goroutine.
type StateStore interface {
	// LoadNextRun returns the persisted next-run time for jobName. found
	// is false if the store has nothing for this job yet (first run ever,
	// a new job, or a store that was cleared) — the Scheduler falls back
	// to computing the schedule fresh in that case; it is not an error.
	//
	// A returned next time in the past (the process was down past its due
	// time) is expected and handled by the Scheduler as a catch-up run:
	// it fires almost immediately, then reschedules normally.
	LoadNextRun(ctx context.Context, jobName string) (next time.Time, found bool, err error)

	// SaveNextRun persists jobName's next scheduled run time. Called after
	// every completed Execute() — whether the job ultimately succeeded or
	// exhausted its retries — and once more right after the very first
	// schedule computation, so a crash before any run still resumes
	// correctly on the next restart.
	SaveNextRun(ctx context.Context, jobName string, next time.Time) error
}
