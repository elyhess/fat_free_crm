package service

import (
	"context"
	"log/slog"
	"time"
)

// Job represents a periodic task.
type Job struct {
	Name     string
	Interval time.Duration
	Fn       func() error
}

// Scheduler runs periodic jobs in the background.
type Scheduler struct {
	jobs   []Job
	cancel context.CancelFunc
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// Add registers a job to run at the given interval.
func (s *Scheduler) Add(name string, interval time.Duration, fn func() error) {
	s.jobs = append(s.jobs, Job{Name: name, Interval: interval, Fn: fn})
}

// Start launches all jobs as goroutines. Call Stop() to shut down.
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	for _, job := range s.jobs {
		go s.runJob(ctx, job)
	}

	slog.Info("scheduler started", "jobs", len(s.jobs))
}

// Stop signals all jobs to stop.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Scheduler) runJob(ctx context.Context, job Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run once immediately
	s.executeJob(job)

	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler: job stopped", "name", job.Name)
			return
		case <-ticker.C:
			s.executeJob(job)
		}
	}
}

func (s *Scheduler) executeJob(job Job) {
	start := time.Now()
	if err := job.Fn(); err != nil {
		slog.Error("scheduler: job failed", "name", job.Name, "error", err, "duration", time.Since(start))
	} else {
		slog.Info("scheduler: job completed", "name", job.Name, "duration", time.Since(start))
	}
}
