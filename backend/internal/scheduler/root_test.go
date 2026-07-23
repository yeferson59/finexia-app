package scheduler

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// Regression test: a retried attempt must run with a fresh, non-cancelled
// context even when RunnerOptions.Timeout is set. Previously each attempt
// derived its timeout context from the *previous* attempt's already-
// cancelled context, so every retry started with ctx.Done() already closed.
func TestExecute_RetryGetsFreshContext(t *testing.T) {
	var attempts atomic.Int32
	var sawCancelledCtxOnEntry atomic.Bool

	job := JobFunc{
		JobName: "retry-fresh-ctx",
		Fn: func(ctx context.Context) error {
			n := attempts.Add(1)

			select {
			case <-ctx.Done():
				sawCancelledCtxOnEntry.Store(true)
			default:
			}

			if n < 2 {
				return errors.New("boom")
			}

			return nil
		},
	}

	runner := NewRunner(RunnerOptions{
		Timeout:     time.Second,
		MaxRetries:  2,
		BackoffBase: time.Millisecond,
		Log:         logger.Noop(),
	})

	runner.Execute(job)

	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}

	if sawCancelledCtxOnEntry.Load() {
		t.Fatal("job observed an already-cancelled context on a retry attempt")
	}
}

func TestExecute_MaxRetriesNegativeDoesNotPanic(t *testing.T) {
	var attempts atomic.Int32

	job := JobFunc{
		JobName: "negative-retries",
		Fn: func(ctx context.Context) error {
			attempts.Add(1)
			return nil
		},
	}

	runner := NewRunner(RunnerOptions{MaxRetries: -5, Log: logger.Noop()})
	runner.Execute(job)

	if got := attempts.Load(); got != 1 {
		t.Fatalf("expected the job to still run once, got %d attempts", got)
	}
}

func TestExecute_NilLogDoesNotPanic(t *testing.T) {
	job := JobFunc{JobName: "nil-log", Fn: func(ctx context.Context) error { return nil }}

	runner := NewRunner(RunnerOptions{})
	runner.Execute(job)
}

func TestExecute_OnErrorCalledAfterExhaustingRetries(t *testing.T) {
	var attempts atomic.Int32
	var onErrorJob string
	var onErrorErr error

	job := JobFunc{
		JobName: "always-fails",
		Fn: func(ctx context.Context) error {
			attempts.Add(1)
			return errors.New("permanent failure")
		},
	}

	runner := NewRunner(RunnerOptions{
		MaxRetries:  2,
		BackoffBase: time.Millisecond,
		Log:         logger.Noop(),
		OnError: func(jobName string, err error) {
			onErrorJob = jobName
			onErrorErr = err
		},
	})

	runner.Execute(job)

	if got := attempts.Load(); got != 3 {
		t.Fatalf("expected 1 initial attempt + 2 retries = 3, got %d", got)
	}

	if onErrorJob != "always-fails" {
		t.Fatalf("OnError called with job name %q, want %q", onErrorJob, "always-fails")
	}

	if onErrorErr == nil || onErrorErr.Error() != "permanent failure" {
		t.Fatalf("OnError called with err %v, want %q", onErrorErr, "permanent failure")
	}
}

func TestExecute_BackoffMaxCapsDelay(t *testing.T) {
	var timestamps []time.Time
	var mu sync.Mutex

	job := JobFunc{
		JobName: "capped-backoff",
		Fn: func(ctx context.Context) error {
			mu.Lock()
			timestamps = append(timestamps, time.Now())
			mu.Unlock()
			return errors.New("boom")
		},
	}

	runner := NewRunner(RunnerOptions{
		MaxRetries:  2,
		BackoffBase: 30 * time.Millisecond,
		BackoffMax:  20 * time.Millisecond, // lower than BackoffBase: every wait must be capped
		Log:         logger.Noop(),
	})

	start := time.Now()
	runner.Execute(job)
	total := time.Since(start)

	mu.Lock()
	n := len(timestamps)
	mu.Unlock()

	if n != 3 {
		t.Fatalf("expected 3 attempts, got %d", n)
	}

	// Uncapped backoff (30ms + 60ms = 90ms) would clearly exceed this bound;
	// capped backoff (20ms + 20ms = 40ms, plus scheduling slack) should not.
	if total > 80*time.Millisecond {
		t.Fatalf("total duration %v suggests BackoffMax was not applied", total)
	}
}

func TestExecute_SkipsWhenAlreadyRunning(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	var runs atomic.Int32

	job := JobFunc{
		JobName: "slow",
		Fn: func(ctx context.Context) error {
			runs.Add(1)
			close(started)
			<-release
			return nil
		},
	}

	runner := newTestRunner()

	go runner.Execute(job)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first execution to start")
	}

	// A second Execute() while the first is still running must be skipped
	// immediately rather than blocking or running concurrently.
	done := make(chan struct{})
	go func() {
		runner.Execute(job)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second Execute() did not return promptly — tryLock skip failed")
	}

	close(release)

	if got := runs.Load(); got != 1 {
		t.Fatalf("expected the job body to run exactly once, got %d", got)
	}
}

func TestSafeRun_PanicIncludesValue(t *testing.T) {
	runner := NewRunner(RunnerOptions{Log: logger.Noop()})

	job := JobFunc{JobName: "panicky", Fn: func(ctx context.Context) error {
		panic("kaboom")
	}}

	err := runner.safeRun(context.Background(), job)
	if err == nil {
		t.Fatal("expected an error from a panicking job")
	}

	if err.Error() != "panic recuperado: kaboom" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}
