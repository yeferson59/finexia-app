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

	runner.Execute(context.Background(), job)

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
	runner.Execute(context.Background(), job)

	if got := attempts.Load(); got != 1 {
		t.Fatalf("expected the job to still run once, got %d attempts", got)
	}
}

func TestExecute_NilLogDoesNotPanic(t *testing.T) {
	job := JobFunc{JobName: "nil-log", Fn: func(ctx context.Context) error { return nil }}

	runner := NewRunner(RunnerOptions{})
	runner.Execute(context.Background(), job)
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

	runner.Execute(context.Background(), job)

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
	runner.Execute(context.Background(), job)
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

	go runner.Execute(context.Background(), job)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first execution to start")
	}

	// A second Execute() while the first is still running must be skipped
	// immediately rather than blocking or running concurrently.
	done := make(chan struct{})
	go func() {
		runner.Execute(context.Background(), job)
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

func TestExecute_ContextCancelledStopsRetriesEarly(t *testing.T) {
	var attempts atomic.Int32

	job := JobFunc{
		JobName: "cancel-me",
		Fn: func(ctx context.Context) error {
			attempts.Add(1)
			return errors.New("boom")
		},
	}

	runner := NewRunner(RunnerOptions{
		MaxRetries:  10,
		BackoffBase: 200 * time.Millisecond,
		Log:         logger.Noop(),
	})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		runner.Execute(ctx, job)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond) // let the first attempt run and enter backoff
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Execute did not return promptly after ctx cancellation — backoff wait ignored ctx.Done()")
	}

	if got := attempts.Load(); got >= 10 {
		t.Fatalf("expected cancellation to cut retries short, got %d of up to 10 attempts", got)
	}
}

func TestExecute_PerJobOverrideLimitsRetries(t *testing.T) {
	var attempts atomic.Int32

	job := JobFunc{
		JobName: "override-retries",
		Fn: func(ctx context.Context) error {
			attempts.Add(1)
			return errors.New("boom")
		},
	}

	runner := NewRunner(RunnerOptions{MaxRetries: 5, Log: logger.Noop()})

	runner.Execute(context.Background(), job, JobOptions{MaxRetries: Retries(0)})

	if got := attempts.Load(); got != 1 {
		t.Fatalf("expected the override (0 retries) to limit to 1 attempt, got %d", got)
	}
}

func TestExecute_PerJobOverrideBackoffTakesPrecedence(t *testing.T) {
	var timestamps []time.Time
	var mu sync.Mutex

	job := JobFunc{
		JobName: "override-backoff",
		Fn: func(ctx context.Context) error {
			mu.Lock()
			timestamps = append(timestamps, time.Now())
			mu.Unlock()

			return errors.New("boom")
		},
	}

	// The Runner's own default would retry with a slow backoff; the
	// per-job override should be the one that actually applies.
	runner := NewRunner(RunnerOptions{MaxRetries: 1, BackoffBase: time.Second, Log: logger.Noop()})

	start := time.Now()
	runner.Execute(context.Background(), job, JobOptions{MaxRetries: Retries(1), BackoffBase: time.Millisecond})
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Fatalf("override backoff doesn't seem to have been applied: took %v", elapsed)
	}

	mu.Lock()
	n := len(timestamps)
	mu.Unlock()

	if n != 2 {
		t.Fatalf("expected 2 attempts (1 retry), got %d", n)
	}
}

func TestExecute_OnAttemptPanicDoesNotPropagate(t *testing.T) {
	job := JobFunc{JobName: "ok", Fn: func(ctx context.Context) error { return nil }}

	runner := NewRunner(RunnerOptions{
		Log:       logger.Noop(),
		OnAttempt: func(AttemptResult) { panic("boom in OnAttempt") },
	})

	runner.Execute(context.Background(), job) // must not panic
}

func TestExecute_OnErrorPanicDoesNotPropagate(t *testing.T) {
	job := JobFunc{JobName: "always-fails", Fn: func(ctx context.Context) error { return errors.New("boom") }}

	runner := NewRunner(RunnerOptions{
		Log:     logger.Noop(),
		OnError: func(string, error) { panic("boom in OnError") },
	})

	runner.Execute(context.Background(), job) // must not panic
}

func TestBackoffDelay(t *testing.T) {
	tests := []struct {
		name         string
		base, capMax time.Duration
		attempt      int
		want         time.Duration
	}{
		{name: "first attempt equals base", base: time.Second, attempt: 1, want: time.Second},
		{name: "second attempt doubles", base: time.Second, attempt: 2, want: 2 * time.Second},
		{name: "third attempt quadruples", base: time.Second, attempt: 3, want: 4 * time.Second},
		{name: "capped by backoffMax", base: time.Second, capMax: 3 * time.Second, attempt: 5, want: 3 * time.Second},
		{name: "zero base disables backoff", base: 0, capMax: time.Second, attempt: 5, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := backoffDelay(tt.base, tt.capMax, tt.attempt)
			if got != tt.want {
				t.Fatalf("backoffDelay(%v, %v, %d) = %v, want %v", tt.base, tt.capMax, tt.attempt, got, tt.want)
			}
		})
	}
}

func TestBackoffDelay_HighAttemptCountsDoNotOverflow(t *testing.T) {
	// Without overflow protection, doubling time.Second ~63 times wraps
	// time.Duration's int64 range into a negative/garbage value.
	got := backoffDelay(time.Second, 0, 1000)
	if got <= 0 {
		t.Fatalf("backoffDelay overflowed to a non-positive duration: %v", got)
	}

	if got != maxDuration {
		t.Fatalf("expected saturation at maxDuration, got %v", got)
	}
}

func TestBackoffDelay_CapShortCircuitsBeforeOverflowRisk(t *testing.T) {
	got := backoffDelay(time.Second, 10*time.Second, 1000)
	if got != 10*time.Second {
		t.Fatalf("expected the cap to short-circuit the doubling loop, got %v", got)
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
