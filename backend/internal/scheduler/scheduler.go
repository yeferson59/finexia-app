package scheduler

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// Schedule decides when the next run happens, given 'now'. Any scheduling
// strategy implements this single interface. Next is expected to return a
// time; the Scheduler defensively floors the resulting delay at 1ms if
// Next returns a non-future time (e.g. a misconfigured Every{Interval: 0}),
// so a broken Schedule can't spin the loop tightly.
type Schedule interface {
	Next(now time.Time) time.Time
}

// Every fires the job every 'Interval', counted from the moment the previous
// run finishes (not from when it started). Interval must be positive; a
// zero or negative Interval makes the Scheduler re-fire as fast as it can
// (floored at roughly once per millisecond) instead of on any real cadence.
type Every struct {
	Interval time.Duration
}

func (e Every) Next(now time.Time) time.Time {
	return now.Add(e.Interval)
}

// DailyAt fires the job once a day at the given time (local time).
type DailyAt struct {
	Hour   int
	Minute int
}

func (d DailyAt) Next(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), d.Hour, d.Minute, 0, 0, now.Location())

	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}

// Delayed wraps any Schedule and adds a fixed delay to the result of
// Next(). Useful for cases like "every day at 6am, but delayed 10 min"
// without having to touch DailyAt directly (that one case could just add
// the minutes by hand, but Delayed works the same way over Every or any
// future Schedule).
type Delayed struct {
	Schedule Schedule
	Delay    time.Duration
}

func (d Delayed) Next(now time.Time) time.Time {
	return d.Schedule.Next(now).Add(d.Delay)
}

// Jitter wraps a Schedule and adds a random delay between 0 and Max,
// recalculated on every trigger (unlike Delayed's fixed offset). Useful
// when running several instances/replicas of the same process and you
// don't want them all firing the job at the exact same second (avoids
// simultaneous load spikes against the same database, etc).
type Jitter struct {
	Schedule Schedule
	Max      time.Duration
}

func (j Jitter) Next(now time.Time) time.Time {
	if j.Max <= 0 {
		return j.Schedule.Next(now)
	}

	extra := time.Duration(rand.Int63n(int64(j.Max)))

	return j.Schedule.Next(now).Add(extra)
}

// WeeklyAt fires the job once a week, on the given day and time (local
// time). E.g. WeeklyAt{Day: time.Monday, Hour: 8, Minute: 0}.
type WeeklyAt struct {
	Day    time.Weekday
	Hour   int
	Minute int
}

func (w WeeklyAt) Next(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), w.Hour, w.Minute, 0, 0, now.Location())
	daysUntil := (int(w.Day) - int(next.Weekday()) + 7) % 7

	next = next.AddDate(0, 0, daysUntil)
	if !next.After(now) {
		next = next.AddDate(0, 0, 7)
	}

	return next
}

type scheduledJob struct {
	job   Job
	sched Schedule
	opts  []JobOptions
}

// defaultStoreTimeout bounds each StateStore call so a slow or unreachable
// store degrades one job's persistence instead of blocking it indefinitely.
const defaultStoreTimeout = 5 * time.Second

// SchedulerOptions configures optional, cross-cutting Scheduler behavior.
// The zero value keeps today's default: no persistence, every job's
// schedule computed fresh from time.Now() at Start().
type SchedulerOptions struct {
	// Store persists each job's next-run time across restarts. Optional —
	// leave nil to keep the in-memory-only behavior.
	Store StateStore

	// StoreTimeout bounds each Load/Save call to Store. Defaults to 5s if
	// Store is set and this is left zero.
	StoreTimeout time.Duration
}

// Scheduler coordinates multiple jobs each with its own Schedule, without
// depending on any external cron library. Each job runs in its own goroutine.
type Scheduler struct {
	runner *Runner

	store        StateStore
	storeTimeout time.Duration

	mu      sync.Mutex
	jobs    []scheduledJob
	started bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScheduler builds a Scheduler around runner. opts is variadic purely
// for backward compatibility (NewScheduler(runner) keeps working); at most
// the first element is used.
func NewScheduler(runner *Runner, opts ...SchedulerOptions) *Scheduler {
	var o SchedulerOptions
	if len(opts) > 0 {
		o = opts[0]
	}

	if o.Store != nil && o.StoreTimeout <= 0 {
		o.StoreTimeout = defaultStoreTimeout
	}

	return new(Scheduler{runner: runner, store: o.Store, storeTimeout: o.StoreTimeout})
}

// Register adds a job with its schedule, optionally overriding the
// Runner's default retry policy for this job alone (see JobOptions). Must
// be called before Start; calling it afterwards panics rather than
// silently dropping the job, since that's always a caller bug.
func (s *Scheduler) Register(job Job, sched Schedule, overrides ...JobOptions) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		panic("scheduler: Register called after Start")
	}

	s.jobs = append(s.jobs, scheduledJob{job: job, sched: sched, opts: overrides})
}

// Start launches one goroutine per registered job. Calling it more than
// once panics, since that would silently double every job's firing rate.
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		panic("scheduler: Start called more than once")
	}
	s.started = true

	s.ctx, s.cancel = context.WithCancel(context.Background())

	for _, sj := range s.jobs {
		s.wg.Add(1)
		go s.loop(sj.job, sj.sched, sj.opts)
	}
}

// Stop cancels all loops and waits for them to finish cleanly: no new run
// will be scheduled, and any retry backoff currently in progress is cut
// short immediately. It does not forcibly abort a job attempt already in
// flight inside Runner.Execute — that job's own context-awareness (or lack
// of it) determines whether it notices the cancellation and returns early.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()
}

func (s *Scheduler) loop(job Job, sched Schedule, overrides []JobOptions) {
	defer s.wg.Done()

	next := s.loadNext(job, sched)

	for {
		timer := time.NewTimer(s.delayUntil(next))

		select {
		case <-s.ctx.Done():
			timer.Stop()

			return
		case <-timer.C:
			s.runJob(job, overrides)

			next = s.computeNext(job, sched, time.Now())
			s.saveNext(job, next)
		}
	}
}

// delayUntil floors the wait at 1ms so an overdue or misconfigured next
// time (e.g. a persisted run missed while the process was down, or a
// Schedule like Every{Interval: 0}) can't spin the loop tightly — it fires
// almost immediately instead.
func (s *Scheduler) delayUntil(next time.Time) time.Duration {
	return max(time.Millisecond, time.Until(next))
}

// computeNext calls sched.Next(now), recovering a panic by logging it and
// retrying in a second rather than crashing the process.
func (s *Scheduler) computeNext(job Job, sched Schedule, now time.Time) (next time.Time) {
	defer func() {
		if p := recover(); p != nil {
			s.runner.opts.Log.Error(s.ctx, "scheduler: Schedule.Next panicked, retrying in 1s",
				logger.Str("job", safeJobName(job)), logger.Any("panic", p))

			next = now.Add(time.Second)
		}
	}()

	return sched.Next(now)
}

// loadNext resolves the first next-run time for job when its loop starts:
// from the Store if one is configured and has a persisted value, otherwise
// computed fresh — matching the pre-Store behavior. A freshly computed
// value is saved immediately, so a crash before the job's first run still
// resumes correctly on the next restart.
func (s *Scheduler) loadNext(job Job, sched Schedule) time.Time {
	if s.store != nil {
		ctx, cancel := context.WithTimeout(s.ctx, s.storeTimeout)
		next, found, err := s.store.LoadNextRun(ctx, job.Name())
		cancel()

		switch {
		case err != nil:
			s.runner.opts.Log.Error(s.ctx, "scheduler: failed to load persisted next-run, computing fresh",
				logger.Str("job", safeJobName(job)), logger.Err(err))
		case found:
			return next
		}
	}

	next := s.computeNext(job, sched, time.Now())
	s.saveNext(job, next)

	return next
}

// saveNext persists next for job if a Store is configured. Failures are
// logged, not returned — persistence is best-effort and must never block
// or break scheduling.
func (s *Scheduler) saveNext(job Job, next time.Time) {
	if s.store == nil {
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.storeTimeout)
	defer cancel()

	if err := s.store.SaveNextRun(ctx, job.Name(), next); err != nil {
		s.runner.opts.Log.Error(s.ctx, "scheduler: failed to persist next-run",
			logger.Str("job", safeJobName(job)), logger.Err(err))
	}
}

// runJob executes job through the Runner, recovering any panic that
// escapes Execute (e.g. from a Job whose Name() panics) so one broken job
// can't take down the whole process — the other scheduled jobs keep running.
func (s *Scheduler) runJob(job Job, overrides []JobOptions) {
	defer func() {
		if p := recover(); p != nil {
			s.runner.opts.Log.Error(s.ctx, "scheduler: recovered panic running job",
				logger.Str("job", safeJobName(job)), logger.Any("panic", p))
		}
	}()

	s.runner.Execute(s.ctx, job, overrides...)
}

// safeJobName reads job.Name(), recovering if it panics — used from panic
// handlers themselves, where a Name() that panics (the very thing that may
// have triggered the handler) must not cause a second, unrecovered panic.
func safeJobName(job Job) (name string) {
	defer func() {
		if recover() != nil {
			name = "<unknown: Name() panicked>"
		}
	}()

	return job.Name()
}

//
// runner := NewRunner(RunnerOptions{
// 	Timeout:     30 * time.Second,
// 	MaxRetries:  2,
// 	BackoffBase: 500 * time.Millisecond,
// })
//
// // Without a Store: every job's cadence resets on restart — an
// // Every{Interval: 3 * time.Hour} job starts counting 3h again from
// // whenever the process came back up.
// sched := NewScheduler(runner)
//
// // With a Store (any backing that implements StateStore — Postgres,
// // Redis, a file, ...): each job resumes from its persisted next-run
// // time instead, surviving restarts/deploys.
// sched = NewScheduler(runner, SchedulerOptions{Store: myStateStore})
//
// sched.Register(JobFunc{JobName: "sync-prices", Fn: syncPrices}, Every{Interval: time.Hour})
//
// // 6:00am + a fixed 10-minute delay -> runs at 6:10am
// reportSchedule := Delayed{Schedule: DailyAt{Hour: 6, Minute: 0}, Delay: 10 * time.Minute}
// sched.Register(JobFunc{JobName: "daily-report", Fn: buildReport}, reportSchedule)
//
// // every hour + up to 2 minutes of random jitter (for multiple replicas)
// jitterSchedule := Jitter{Schedule: Every{Interval: time.Hour}, Max: 2 * time.Minute}
// sched.Register(JobFunc{JobName: "cache-refresh", Fn: refreshCache}, jitterSchedule)
//
// sched.Start()
//
// // on app shutdown:
// sched.Stop()
