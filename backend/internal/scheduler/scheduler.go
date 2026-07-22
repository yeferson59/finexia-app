package scheduler

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Schedule decide cuándo es la próxima ejecución a partir de 'now'.
// Cualquier estrategia de scheduling implementa esta única interfaz.
type Schedule interface {
	Next(now time.Time) time.Time
}

// Every dispara el job cada 'Interval', contado desde el momento en que
// termina de calcularse (no desde que arrancó la ejecución anterior).
type Every struct {
	Interval time.Duration
}

func (e Every) Next(now time.Time) time.Time {
	return now.Add(e.Interval)
}

// DailyAt dispara el job una vez al día a la hora indicada (hora local).
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

// Delayed envuelve cualquier Schedule y le suma un retraso fijo al
// resultado de Next(). Sirve para casos como "todos los días a las
// 6am, pero retrasado 10 min" sin tener que tocar DailyAt directamente
// (en ese caso puntual bastaría con sumar los minutos a mano, pero
// Delayed funciona igual sobre Every o cualquier Schedule futuro).
type Delayed struct {
	Schedule Schedule
	Delay    time.Duration
}

func (d Delayed) Next(now time.Time) time.Time {
	return d.Schedule.Next(now).Add(d.Delay)
}

// Jitter envuelve un Schedule y le suma un retraso aleatorio entre
// 0 y Max, recalculado en cada disparo (no es fijo como Delayed).
// Útil cuando corres varias instancias/réplicas del mismo proceso y
// no quieres que todas ejecuten el job exactamente al mismo segundo
// (evita picos de carga simultáneos contra la misma base de datos, etc).
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

// WeeklyAt dispara el job una vez por semana, en el día y hora indicados
// (hora local). Ej: WeeklyAt{Day: time.Monday, Hour: 8, Minute: 0}.
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

// Scheduler coordina múltiples jobs con su propio Schedule, sin depender
// de ninguna librería externa de cron. Cada job corre en su propia goroutine.
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

// Register agrega un job con su schedule. Debe llamarse antes de Start.
func (s *Scheduler) Register(job Job, sched Schedule) {
	s.jobs = append(s.jobs, scheduledJob{job: job, sched: sched})
}

// Start lanza una goroutine por job registrado.
func (s *Scheduler) Start() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for _, sj := range s.jobs {
		s.wg.Add(1)
		go s.loop(sj.job, sj.sched)
	}
}

// Stop cancela todos los loops y espera a que terminen limpiamente.
// No aborta un job que ya está corriendo dentro de Runner.Execute;
// solo evita que se dispare una próxima ejecución.
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
// // 6:00am + 10 minutos de retraso fijo -> corre a las 6:10am
// reportSchedule := Delayed{Schedule: DailyAt{Hour: 6, Minute: 0}, Delay: 10 * time.Minute}
// sched.Register(JobFunc{JobName: "daily-report", Fn: buildReport}, reportSchedule)
//
// // cada hora + hasta 2 minutos de jitter aleatorio (para varias réplicas)
// jitterSchedule := Jitter{Schedule: Every{Interval: time.Hour}, Max: 2 * time.Minute}
// sched.Register(JobFunc{JobName: "cache-refresh", Fn: refreshCache}, jitterSchedule)
//
// sched.Start()
//
// // en el shutdown de la app:
// sched.Stop()
