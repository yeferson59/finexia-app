package scheduler

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Schedule decides when the next run happens, given 'now'.
// Any scheduling strategy implements this single interface.
type Schedule interface {
	Next(now time.Time) time.Time
}

// Every fires the job every 'Interval', counted from the moment the previous
// run finishes (not from when it started).
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
}

// Scheduler coordinates multiple jobs each with its own Schedule, without
// depending on any external cron library. Each job runs in its own goroutine.
type Scheduler struct {
	runner *Runner
	jobs   []scheduledJob
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewScheduler(runner *Runner) *Scheduler {
	return new(Scheduler{runner: runner})
}

// Register adds a job with its schedule. Must be called before Start.
func (s *Scheduler) Register(job Job, sched Schedule) {
	s.jobs = append(s.jobs, scheduledJob{job: job, sched: sched})
}

// Start launches one goroutine per registered job.
func (s *Scheduler) Start() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for _, sj := range s.jobs {
		s.wg.Add(1)
		go s.loop(sj.job, sj.sched)
	}
}

// Stop cancels all loops and waits for them to finish cleanly. It does
// not abort a job already running inside Runner.Execute; it only
// prevents the next run from firing.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()
}

func (s *Scheduler) loop(job Job, sched Schedule) {
	defer s.wg.Done()

	for {
		now := time.Now()
		next := sched.Next(now)
		delay := max(0, next.Sub(now))
		timer := time.NewTimer(delay)

		select {
		case <-s.ctx.Done():
			timer.Stop()

			return
		case <-timer.C:
			s.runner.Execute(job)
		}
	}
}

//
// runner := NewRunner(RunnerOptions{
// 	Timeout:     30 * time.Second,
// 	MaxRetries:  2,
// 	BackoffBase: 500 * time.Millisecond,
// })
//
// sched := NewScheduler(runner)
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
