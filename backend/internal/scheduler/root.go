package scheduler

import (
	"context"
	"fmt"
	"math"
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

func (r *Runner) wait(attempt int) {
	if r.opts.BackoffBase <= 0 {
		return
	}

	delay := r.opts.BackoffBase * time.Duration(math.Pow(2, float64(attempt-1)))
	if r.opts.BackoffMax > 0 && delay > r.opts.BackoffMax {
		delay = r.opts.BackoffMax
	}

	time.Sleep(delay)
}

func (r *Runner) safeRun(ctx context.Context, job Job) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic recuperado: %v", p)
		}
	}()

	return job.Run(ctx)
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

func (r *Runner) Execute(job Job) {
	base := context.Background()

	if !r.tryLock(job.Name()) {
		r.opts.Log.Info(base, "skip: ya está en ejecución", logger.Str("job", job.Name()))

		return
	}
	defer r.unlock(job.Name())

	totalAttempts := r.opts.MaxRetries + 1
	var lastErr error

	for attempt := 1; attempt <= totalAttempts; attempt++ {
		ctx := base
		var cancel context.CancelFunc

		if r.opts.Timeout > 0 {
			ctx, cancel = context.WithTimeout(base, r.opts.Timeout)
		}

		start, err := time.Now(), r.safeRun(ctx, job)
		duration := time.Since(start)

		if cancel != nil {
			cancel()
		}

		if r.opts.OnAttempt != nil {
			r.opts.OnAttempt(AttemptResult{
				JobName:  job.Name(),
				Attempt:  attempt,
				Duration: duration,
				Err:      err,
			})
		}

		if err == nil {
			r.opts.Log.Info(ctx, "job "+job.Name()+" OK en intento "+strconv.Itoa(attempt)+"/"+strconv.Itoa(totalAttempts)+" ("+duration.String()+")")

			return
		}

		lastErr = err

		r.opts.Log.Error(ctx, "job "+job.Name()+" falló intento "+strconv.Itoa(attempt)+"/"+strconv.Itoa(totalAttempts)+" ("+duration.String()+"): "+err.Error())

		if attempt < totalAttempts {
			r.wait(attempt)
		}
	}

	if lastErr != nil && r.opts.OnError != nil {
		r.opts.OnError(job.Name(), lastErr)
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
// c := cron.New()
// syncJob := JobFunc{JobName: "sync-prices", Fn: syncPrices}
// c.AddFunc("@every 1h", func() { runner.Execute(syncJob) })
// c.Start()
