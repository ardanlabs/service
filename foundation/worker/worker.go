// Package worker manages a set of registered jobs that execute on demand.
package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// JobFunc defines a function that can execute work for a specific job.
type JobFunc func(ctx context.Context, traceID string, payload interface{})

// Worker manages jobs and the execution of those jobs concurrently.
type Worker struct {
	wg       sync.WaitGroup
	mu       sync.Mutex
	registry map[string]JobFunc
	running  map[string]context.CancelFunc
}

// New constructs a Worker for managing and executing jobs.
func New(registry map[string]JobFunc) *Worker {
	return &Worker{
		registry: registry,
		running:  make(map[string]context.CancelFunc),
	}
}

// Running returns the number of jobs running.
func (w *Worker) Running() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return len(w.running)
}

// Shutdown waits for all jobs to complete before it returns.
func (w *Worker) Shutdown(ctx context.Context) error {

	// Call the cancel function for all running goroutines.
	w.mu.Lock()
	{
		for _, cancel := range w.running {
			cancel()
		}
	}
	w.mu.Unlock()

	// Launch a goroutine to wait for all the worker goroutines
	// to complete their work.
	ch := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(ch)
	}()

	// Wait for the goroutines to report they are done or when
	// the timeout is reached.
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Start lookups a job by key and launches a goroutine to perform the work. A
// work key is returned so the caller can cancel work early.
func (w *Worker) Start(ctx context.Context, traceID string, jobKey string, payload interface{}) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Locate the job in the jobs registry.
	f, exists := w.registry[jobKey]
	if !exists {
		return "", fmt.Errorf("job %s is not registered", jobKey)
	}

	// Need a unique key for this work.
	workKey := uuid.NewString()

	// Create a cancel function and keep it for stop/shutdown purposes.
	ctx, cancel := context.WithCancel(ctx)
	w.running[workKey] = cancel

	// Launch a goroutine to perform the work.
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() { cancel(); w.removeWork(workKey) }()
		f(ctx, traceID, payload)
	}()

	return workKey, nil
}

// Stop is used to cancel an existing job that is running.
func (w *Worker) Stop(workKey string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Locate the work in the running list.
	cancel, exists := w.running[workKey]
	if !exists {
		return fmt.Errorf("work %s is not running", workKey)
	}

	// Call cancel to stop the work.
	cancel()

	return nil
}

// Convenience function to remove work from the running list.
func (w *Worker) removeWork(workKey string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.running, workKey)
}
