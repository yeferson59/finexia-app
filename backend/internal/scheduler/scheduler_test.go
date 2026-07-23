package scheduler

import (
	"context"
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
