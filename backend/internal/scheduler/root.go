package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type Job interface {
	Name() string
	Run(ctx context.Context) error
}

type JobFunc struct {
	JobName string
	Fn      func(ctx context.Context) error
}

func (j JobFunc) Name() string {
	return j.JobName
}

func (j JobFunc) Run(ctx context.Context) error {
	return j.Fn(ctx)
}

type AttemptResult struct {
	JobName  string
	Attempt  int
	Duration time.Duration
	Err      error
}

type RunnerOptions struct {
	Timeout     time.Duration
	MaxRetries  int
	BackoffBase time.Duration
	BackoffMax  time.Duration
	OnError     func(jobName string, err error)
	OnAttempt   func(result AttemptResult)
	Log         logger.Logger
}

// JobOptions overrides the Runner's default retry policy for a single job,
// passed to Execute or Scheduler.Register. Timeout/BackoffBase/BackoffMax
// of zero mean "inherit the Runner's default" (which is itself the same
// convention RunnerOptions uses for "disabled"). MaxRetries is a pointer
// because 0 is a meaningful, distinct override ("no retries") from "not
// set" — use the Retries helper to build it.
type JobOptions struct {
	Timeout     time.Duration
	MaxRetries  *int
	BackoffBase time.Duration
	BackoffMax  time.Duration
}

// Retries returns a pointer to n, for JobOptions.MaxRetries — Go has no
// address-of-literal operator.
func Retries(n int) *int {
	return new(n)
}

type Runner struct {
	opts    RunnerOptions
	mu      sync.Mutex
	running map[string]bool
}

func NewRunner(opts RunnerOptions) *Runner {
	if opts.MaxRetries < 0 {
		opts.MaxRetries = 0
	}

	if opts.Log == nil {
		opts.Log = logger.Noop()
	}

	return new(Runner{opts: opts, running: make(map[string]bool)})
}

// effectiveOptions is RunnerOptions' tunable subset after merging in a
// per-job override. OnError/OnAttempt/Log stay global — only the
// retry/timeout policy can be overridden per job.
type effectiveOptions struct {
	Timeout     time.Duration
	MaxRetries  int
	BackoffBase time.Duration
	BackoffMax  time.Duration
}

func (r *Runner) resolveOptions(overrides ...JobOptions) effectiveOptions {
	eo := effectiveOptions{
		Timeout:     r.opts.Timeout,
		MaxRetries:  r.opts.MaxRetries,
		BackoffBase: r.opts.BackoffBase,
		BackoffMax:  r.opts.BackoffMax,
	}

	if len(overrides) == 0 {
		return eo
	}

	o := overrides[0]

	if o.Timeout > 0 {
		eo.Timeout = o.Timeout
	}

	if o.MaxRetries != nil {
		eo.MaxRetries = max(*o.MaxRetries, 0)
	}

	if o.BackoffBase > 0 {
		eo.BackoffBase = o.BackoffBase
	}

	if o.BackoffMax > 0 {
		eo.BackoffMax = o.BackoffMax
	}

	return eo
}

// maxDuration is the largest representable time.Duration, used to keep
// backoffDelay's doubling from overflowing int64 when MaxRetries is large.
const maxDuration = time.Duration(1<<63 - 1)

// backoffDelay computes the exponential backoff for the given attempt
// (1-indexed: attempt 1 -> base, attempt 2 -> 2*base, ...), saturating at
// backoffMax (if set) and never overflowing time.Duration's int64 range.
func backoffDelay(base, backoffMax time.Duration, attempt int) time.Duration {
	if base <= 0 {
		return 0
	}

	delay := base

	for i := 1; i < attempt; i++ {
		if backoffMax > 0 && delay >= backoffMax {
			return backoffMax
		}

		if delay > maxDuration/2 {
			delay = maxDuration

			break
		}

		delay *= 2
	}

	if backoffMax > 0 && delay > backoffMax {
		delay = backoffMax
	}

	return delay
}

// wait sleeps for the backoff delay of the given attempt, or returns early
// if ctx is cancelled — so a Scheduler.Stop() during a retry backoff
// doesn't have to wait out the full delay.
func (r *Runner) wait(ctx context.Context, backoffBase, backoffMax time.Duration, attempt int) {
	delay := backoffDelay(backoffBase, backoffMax, attempt)
	if delay <= 0 {
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (r *Runner) safeRun(ctx context.Context, job Job) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic recuperado: %v", p)
		}
	}()

	return job.Run(ctx)
}

// safeCallback runs fn, recovering and logging any panic instead of
// letting it crash the calling goroutine (and, since these run inside
// goroutines spawned by Scheduler, potentially the whole process).
func (r *Runner) safeCallback(ctx context.Context, name string, fn func()) {
	defer func() {
		if p := recover(); p != nil {
			r.opts.Log.Error(ctx, "scheduler: "+name+" callback panicked", logger.Any("panic", p))
		}
	}()

	fn()
}

func (r *Runner) tryLock(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running[name] {
		return false
	}

	r.running[name] = true

	return true
}

func (r *Runner) unlock(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.running[name] = false
}

// Execute runs job to completion, retrying on failure per the Runner's
// options (or the given per-job override). ctx bounds the whole call: if
// it's cancelled, Execute stops retrying (and interrupts any backoff wait
// in progress) as soon as the current attempt returns, without wasting a
// further attempt on an already-dead context.
func (r *Runner) Execute(ctx context.Context, job Job, overrides ...JobOptions) {
	if !r.tryLock(job.Name()) {
		r.opts.Log.Info(ctx, "skip: ya está en ejecución", logger.Str("job", job.Name()))

		return
	}
	defer r.unlock(job.Name())

	opts := r.resolveOptions(overrides...)
	totalAttempts := opts.MaxRetries + 1
	var lastErr error

	for attempt := 1; attempt <= totalAttempts; attempt++ {
		if cErr := ctx.Err(); cErr != nil {
			lastErr = cErr

			break
		}

		attemptCtx := ctx
		var cancel context.CancelFunc

		if opts.Timeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		}

		start, err := time.Now(), r.safeRun(attemptCtx, job)
		duration := time.Since(start)

		if cancel != nil {
			cancel()
		}

		if r.opts.OnAttempt != nil {
			result := AttemptResult{JobName: job.Name(), Attempt: attempt, Duration: duration, Err: err}
			r.safeCallback(ctx, "OnAttempt", func() { r.opts.OnAttempt(result) })
		}

		if err == nil {
			r.opts.Log.Info(ctx, "job "+job.Name()+" OK en intento "+strconv.Itoa(attempt)+"/"+strconv.Itoa(totalAttempts)+" ("+duration.String()+")")

			return
		}

		lastErr = err

		r.opts.Log.Error(ctx, "job "+job.Name()+" falló intento "+strconv.Itoa(attempt)+"/"+strconv.Itoa(totalAttempts)+" ("+duration.String()+"): "+err.Error())

		if attempt < totalAttempts {
			r.wait(ctx, opts.BackoffBase, opts.BackoffMax, attempt)
		}
	}

	if lastErr != nil && r.opts.OnError != nil {
		r.safeCallback(ctx, "OnError", func() { r.opts.OnError(job.Name(), lastErr) })
	}
}

//
// runner := NewRunner(RunnerOptions{
// 	Timeout:     30 * time.Second,
// 	MaxRetries:  3,
// 	BackoffBase: 500 * time.Millisecond,
// 	BackoffMax:  10 * time.Second,
// 	OnAttempt: func(res AttemptResult) {
// 		metrics.ObserveJobDuration(res.JobName, res.Attempt, res.Duration, res.Err == nil)
// 	},
// 	OnError: func(name string, err error) {
// 		log.Printf("ALERT: job %s failed permanently: %v", name, err)
// 	},
// })
//
// // one job with the Runner's default policy, one that never retries:
// syncJob := JobFunc{JobName: "sync-prices", Fn: syncPrices}
// cleanupJob := JobFunc{JobName: "cleanup", Fn: cleanup}
//
// sched := NewScheduler(runner)
// sched.Register(syncJob, Every{Interval: time.Hour})
// sched.Register(cleanupJob, Every{Interval: 24 * time.Hour}, JobOptions{MaxRetries: Retries(0)})
// sched.Start()
//
// // on shutdown: cancels pending retries/backoff waits and stops scheduling
// // new runs; does not abort a job attempt already in flight unless that
// // job itself observes ctx cancellation.
// sched.Stop()
//
