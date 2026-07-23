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

// fixedNow is a Wednesday, used as the reference point for every Schedule
// test so the expected values below are reproducible regardless of when
// the test suite runs.
var fixedNow = time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)

func TestEvery_Next(t *testing.T) {
	e := Every{Interval: 90 * time.Minute}

	got := e.Next(fixedNow)
	want := fixedNow.Add(90 * time.Minute)

	if !got.Equal(want) {
		t.Fatalf("Every.Next() = %v, want %v", got, want)
	}
}

func TestDailyAt_Next(t *testing.T) {
	tests := []struct {
		name string
		d    DailyAt
		want time.Time
	}{
		{
			name: "target later today stays today",
			d:    DailyAt{Hour: 11, Minute: 0},
			want: time.Date(2026, 7, 22, 11, 0, 0, 0, time.UTC),
		},
		{
			name: "target already passed today rolls to tomorrow",
			d:    DailyAt{Hour: 9, Minute: 0},
			want: time.Date(2026, 7, 23, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "target exactly equal to now rolls to tomorrow",
			d:    DailyAt{Hour: 10, Minute: 0},
			want: time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.Next(fixedNow)
			if !got.Equal(tt.want) {
				t.Fatalf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeeklyAt_Next(t *testing.T) {
	tests := []struct {
		name string
		w    WeeklyAt
		want time.Time
	}{
		{
			name: "same weekday, time still ahead today",
			w:    WeeklyAt{Day: time.Wednesday, Hour: 11, Minute: 0},
			want: time.Date(2026, 7, 22, 11, 0, 0, 0, time.UTC),
		},
		{
			name: "same weekday, time already passed today rolls a full week",
			w:    WeeklyAt{Day: time.Wednesday, Hour: 9, Minute: 0},
			want: time.Date(2026, 7, 29, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "later weekday this week",
			w:    WeeklyAt{Day: time.Friday, Hour: 9, Minute: 0},
			want: time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "earlier weekday wraps to next week",
			w:    WeeklyAt{Day: time.Monday, Hour: 9, Minute: 0},
			want: time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.w.Next(fixedNow)
			if !got.Equal(tt.want) {
				t.Fatalf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDelayed_Next(t *testing.T) {
	d := Delayed{Schedule: Every{Interval: time.Hour}, Delay: 10 * time.Minute}

	got := d.Next(fixedNow)
	want := fixedNow.Add(time.Hour).Add(10 * time.Minute)

	if !got.Equal(want) {
		t.Fatalf("Delayed.Next() = %v, want %v", got, want)
	}
}

func TestJitter_Next(t *testing.T) {
	t.Run("zero max passes through unchanged", func(t *testing.T) {
		j := Jitter{Schedule: Every{Interval: time.Hour}, Max: 0}

		got := j.Next(fixedNow)
		want := fixedNow.Add(time.Hour)

		if !got.Equal(want) {
			t.Fatalf("Next() = %v, want %v", got, want)
		}
	})

	t.Run("positive max stays within [base, base+max]", func(t *testing.T) {
		j := Jitter{Schedule: Every{Interval: time.Hour}, Max: 2 * time.Minute}
		base := fixedNow.Add(time.Hour)

		for range 50 {
			got := j.Next(fixedNow)
			if got.Before(base) || got.After(base.Add(2*time.Minute)) {
				t.Fatalf("Next() = %v, want within [%v, %v]", got, base, base.Add(2*time.Minute))
			}
		}
	})
}

func newTestRunner() *Runner {
	return NewRunner(RunnerOptions{Log: logger.Noop()})
}

// tick fires a Schedule that always triggers as soon as possible after the
// previous run, letting Scheduler.loop run many times within a short test.
type tick struct{ interval time.Duration }

func (t tick) Next(now time.Time) time.Time {
	return now.Add(t.interval)
}

// once fires immediately on its first Next() call and then far in the
// future, so a job scheduled with it runs exactly once during a short
// test, isolating assertions about retries from repeated schedule triggers.
type once struct{ fired bool }

func (o *once) Next(now time.Time) time.Time {
	if o.fired {
		return now.Add(24 * time.Hour)
	}

	o.fired = true

	return now
}

func TestScheduler_StartExecutesRegisteredJobs(t *testing.T) {
	sched := NewScheduler(newTestRunner())

	fired := make(chan struct{}, 8)
	job := JobFunc{
		JobName: "tick",
		Fn: func(ctx context.Context) error {
			fired <- struct{}{}
			return nil
		},
	}

	sched.Register(job, tick{interval: 5 * time.Millisecond})
	sched.Start()
	defer sched.Stop()

	for range 3 {
		select {
		case <-fired:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for job execution")
		}
	}
}

func TestScheduler_StopHaltsFurtherExecutions(t *testing.T) {
	sched := NewScheduler(newTestRunner())

	fired := make(chan struct{}, 32)
	job := JobFunc{
		JobName: "tick",
		Fn: func(ctx context.Context) error {
			fired <- struct{}{}
			return nil
		},
	}

	sched.Register(job, tick{interval: 5 * time.Millisecond})
	sched.Start()

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first job execution")
	}

	stopped := make(chan struct{})
	go func() {
		sched.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop() did not return in time — loop goroutine leaked")
	}

	// Drain whatever was already in flight when Stop() was called, then make
	// sure nothing new arrives afterwards.
	drain := true
	for drain {
		select {
		case <-fired:
		default:
			drain = false
		}
	}

	select {
	case <-fired:
		t.Fatal("job fired again after Stop()")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestScheduler_StopWithoutStartDoesNotBlock(t *testing.T) {
	sched := NewScheduler(newTestRunner())
	sched.Register(JobFunc{JobName: "noop", Fn: func(ctx context.Context) error { return nil }}, Every{Interval: time.Hour})

	done := make(chan struct{})
	go func() {
		sched.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop() blocked when Start() was never called")
	}
}

func TestScheduler_MultipleJobsRunIndependently(t *testing.T) {
	sched := NewScheduler(newTestRunner())

	firedA := make(chan struct{}, 8)
	firedB := make(chan struct{}, 8)

	sched.Register(JobFunc{JobName: "a", Fn: func(ctx context.Context) error { firedA <- struct{}{}; return nil }}, tick{interval: 5 * time.Millisecond})
	sched.Register(JobFunc{JobName: "b", Fn: func(ctx context.Context) error { firedB <- struct{}{}; return nil }}, tick{interval: 5 * time.Millisecond})

	sched.Start()
	defer sched.Stop()

	for _, ch := range []chan struct{}{firedA, firedB} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for one of the two independent jobs")
		}
	}
}

func TestScheduler_StartCalledTwicePanics(t *testing.T) {
	sched := NewScheduler(newTestRunner())
	sched.Start()
	defer sched.Stop()

	defer func() {
		if recover() == nil {
			t.Fatal("expected a second Start() call to panic")
		}
	}()

	sched.Start()
}

func TestScheduler_RegisterAfterStartPanics(t *testing.T) {
	sched := NewScheduler(newTestRunner())
	sched.Start()
	defer sched.Stop()

	defer func() {
		if recover() == nil {
			t.Fatal("expected Register() after Start() to panic")
		}
	}()

	sched.Register(JobFunc{JobName: "late", Fn: func(ctx context.Context) error { return nil }}, Every{Interval: time.Hour})
}

func TestScheduler_RegisterWithPerJobOverride(t *testing.T) {
	runner := NewRunner(RunnerOptions{MaxRetries: 5, BackoffBase: time.Second, Log: logger.Noop()})
	sched := NewScheduler(runner)

	var attempts atomic.Int32
	firstAttempt := make(chan struct{})

	job := JobFunc{
		JobName: "override-sched",
		Fn: func(ctx context.Context) error {
			if attempts.Add(1) == 1 {
				close(firstAttempt)
			}

			return errors.New("boom")
		},
	}

	// &once{} fires immediately, then never again during this test;
	// isolates the assertion to the retry override, not repeated triggers.
	sched.Register(job, &once{}, JobOptions{MaxRetries: Retries(0)})
	sched.Start()
	defer sched.Stop()

	select {
	case <-firstAttempt:
	case <-time.After(time.Second):
		t.Fatal("job never fired")
	}

	time.Sleep(50 * time.Millisecond) // give an incorrect retry a chance to fire

	if got := attempts.Load(); got != 1 {
		t.Fatalf("expected the per-job MaxRetries override (0) to prevent retries, got %d attempts", got)
	}
}

// panickyNameJob's Name() panics, exercising the panic surface inside
// Runner.Execute that isn't covered by safeRun (which only wraps job.Run).
type panickyNameJob struct{}

func (panickyNameJob) Name() string                { panic("name panic") }
func (panickyNameJob) Run(_ context.Context) error { return nil }

func TestScheduler_PanicInOneJobDoesNotStopOthers(t *testing.T) {
	sched := NewScheduler(newTestRunner())

	fired := make(chan struct{}, 8)
	goodJob := JobFunc{JobName: "good", Fn: func(ctx context.Context) error { fired <- struct{}{}; return nil }}

	sched.Register(panickyNameJob{}, tick{interval: 5 * time.Millisecond})
	sched.Register(goodJob, tick{interval: 5 * time.Millisecond})
	sched.Start()
	defer sched.Stop()

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("good job never fired — a panic in the other job likely crashed its goroutine (or the process)")
	}
}

func TestScheduler_DelayUntilFloorsNonPositiveDelay(t *testing.T) {
	sched := NewScheduler(newTestRunner())

	if got := sched.delayUntil(time.Now()); got != time.Millisecond {
		t.Fatalf("expected a 1ms floor for a 'now' target, got %v", got)
	}

	if got := sched.delayUntil(time.Now().Add(-time.Hour)); got != time.Millisecond {
		t.Fatalf("expected a 1ms floor for a past target, got %v", got)
	}
}

type panickyNextSchedule struct{}

func (panickyNextSchedule) Next(time.Time) time.Time { panic("next panic") }

func TestScheduler_ComputeNextRecoversPanickingSchedule(t *testing.T) {
	sched := NewScheduler(newTestRunner())
	sched.ctx = context.Background()

	job := JobFunc{JobName: "panicky-next", Fn: func(ctx context.Context) error { return nil }}
	now := time.Now()

	got := sched.computeNext(job, panickyNextSchedule{}, now)
	if !got.After(now) {
		t.Fatalf("expected computeNext to fall back to a future time after a panic, got %v (now=%v)", got, now)
	}
}

func TestScheduler_StopInterruptsJobRetryBackoff(t *testing.T) {
	runner := NewRunner(RunnerOptions{
		MaxRetries:  20,
		BackoffBase: time.Second, // deliberately long
		Log:         logger.Noop(),
	})
	sched := NewScheduler(runner)

	started := make(chan struct{})
	var startedOnce sync.Once

	job := JobFunc{
		JobName: "slow-retry",
		Fn: func(ctx context.Context) error {
			startedOnce.Do(func() { close(started) })
			return errors.New("boom")
		},
	}

	sched.Register(job, &once{})
	sched.Start()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job never started")
	}

	time.Sleep(20 * time.Millisecond) // let it enter the long backoff sleep

	stopped := make(chan struct{})
	go func() {
		sched.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop() did not return promptly — cancellation didn't reach the retry backoff")
	}
}

// fakeStore is an in-memory StateStore for tests, standing in for a
// Postgres- or Redis-backed implementation. loadErr/saveErr let tests
// simulate a degraded store independently for each operation.
type fakeStore struct {
	mu      sync.Mutex
	data    map[string]time.Time
	loadErr error
	saveErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{data: make(map[string]time.Time)}
}

func (f *fakeStore) LoadNextRun(_ context.Context, jobName string) (time.Time, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.loadErr != nil {
		return time.Time{}, false, f.loadErr
	}

	next, ok := f.data[jobName]

	return next, ok, nil
}

func (f *fakeStore) SaveNextRun(_ context.Context, jobName string, next time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.saveErr != nil {
		return f.saveErr
	}

	f.data[jobName] = next

	return nil
}

func (f *fakeStore) get(jobName string) (time.Time, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	next, ok := f.data[jobName]

	return next, ok
}

func TestScheduler_PersistsNextRunAfterExecute(t *testing.T) {
	store := newFakeStore()
	sched := NewScheduler(newTestRunner(), SchedulerOptions{Store: store})

	fired := make(chan struct{}, 1)
	job := JobFunc{JobName: "persist-me", Fn: func(ctx context.Context) error { fired <- struct{}{}; return nil }}

	sched.Register(job, &once{}) // fires immediately
	sched.Start()
	defer sched.Stop()

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("job never fired")
	}

	// The Store may transiently hold the pre-run "next" value saved by
	// loadNext (which for &once{} is "now", not future) until the
	// post-run save overwrites it; poll until it settles on the real,
	// future value instead of asserting on the first sighting.
	deadline := time.Now().Add(time.Second)
	for {
		if next, ok := store.get("persist-me"); ok && next.After(time.Now()) {
			return
		}

		if time.Now().After(deadline) {
			t.Fatal("expected the Store to eventually hold a future persisted next-run time")
		}

		time.Sleep(time.Millisecond)
	}
}

func TestScheduler_ResumesFromPersistedFutureNextRun(t *testing.T) {
	store := newFakeStore()
	const jobName = "resume-future"
	store.data[jobName] = time.Now().Add(300 * time.Millisecond)

	sched := NewScheduler(newTestRunner(), SchedulerOptions{Store: store})

	fired := make(chan struct{}, 4)
	// tick{interval: 5ms} would fire almost instantly if the persisted
	// next-run weren't consulted at all.
	job := JobFunc{JobName: jobName, Fn: func(ctx context.Context) error { fired <- struct{}{}; return nil }}

	sched.Register(job, tick{interval: 5 * time.Millisecond})
	sched.Start()
	defer sched.Stop()

	select {
	case <-fired:
		t.Fatal("job fired before its persisted next-run time — the Store was not consulted")
	case <-time.After(150 * time.Millisecond):
	}

	select {
	case <-fired:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("job never fired after its persisted next-run time elapsed")
	}
}

func TestScheduler_CatchesUpOverdueRunFromStore(t *testing.T) {
	store := newFakeStore()
	const jobName = "catch-up"
	store.data[jobName] = time.Now().Add(-time.Hour)

	sched := NewScheduler(newTestRunner(), SchedulerOptions{Store: store})

	fired := make(chan struct{}, 1)
	job := JobFunc{JobName: jobName, Fn: func(ctx context.Context) error { fired <- struct{}{}; return nil }}

	// Every{Interval: time.Hour} would otherwise not fire for an hour; the
	// overdue persisted next-run should trigger an immediate catch-up run.
	sched.Register(job, Every{Interval: time.Hour})
	sched.Start()
	defer sched.Stop()

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("overdue persisted job never caught up")
	}
}

func TestScheduler_FallsBackToFreshScheduleOnStoreLoadError(t *testing.T) {
	store := newFakeStore()
	store.loadErr = errors.New("store unavailable")

	sched := NewScheduler(newTestRunner(), SchedulerOptions{Store: store})

	fired := make(chan struct{}, 1)
	job := JobFunc{JobName: "load-error", Fn: func(ctx context.Context) error { fired <- struct{}{}; return nil }}

	sched.Register(job, tick{interval: 5 * time.Millisecond})
	sched.Start()
	defer sched.Stop()

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("job never fired despite a Store load error — expected a fallback to a fresh schedule")
	}
}

func TestScheduler_SaveErrorDoesNotBlockScheduling(t *testing.T) {
	store := newFakeStore()
	store.saveErr = errors.New("store unavailable")

	sched := NewScheduler(newTestRunner(), SchedulerOptions{Store: store})

	fired := make(chan struct{}, 8)
	job := JobFunc{JobName: "save-fails", Fn: func(ctx context.Context) error { fired <- struct{}{}; return nil }}

	sched.Register(job, tick{interval: 5 * time.Millisecond})
	sched.Start()
	defer sched.Stop()

	for range 3 {
		select {
		case <-fired:
		case <-time.After(time.Second):
			t.Fatal("job stopped firing after a Store save error — persistence failures must not block scheduling")
		}
	}
}

func TestNewScheduler_DefaultsStoreTimeoutWhenStoreConfigured(t *testing.T) {
	sched := NewScheduler(newTestRunner(), SchedulerOptions{Store: newFakeStore()})

	if sched.storeTimeout != defaultStoreTimeout {
		t.Fatalf("expected the default store timeout to be applied, got %v", sched.storeTimeout)
	}
}

func TestNewScheduler_NoStoreLeavesTimeoutZero(t *testing.T) {
	sched := NewScheduler(newTestRunner())

	if sched.store != nil {
		t.Fatal("expected no Store by default")
	}
}
