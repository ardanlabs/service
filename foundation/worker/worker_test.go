package worker_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ardanlabs/service/foundation/worker"
)

func Test_Worker(t *testing.T) {

	// Define a work function that waits to be canceled.
	work := func(ctx context.Context) {
		t.Logf("Goroutine running")
		<-ctx.Done()
		t.Logf("Goroutine terminating")
	}

	// Create a worker and start all 4 jobs.
	w, err := worker.New(4)
	if err != nil {
		t.Fatalf("Should be able to create a worker with max 4 : %s", err)
	}

	for i := 0; i < 4; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		if _, err := w.Start(ctx, work); err != nil {
			t.Fatalf("Should be able to execute work : %s", err)
		}
	}

	// Wait for all the jobs to finish.
	for i := 0; i < 4; i++ {
		if w.Running() == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Check that the running list is empty.
	if r := w.Running(); r != 0 {
		t.Errorf("Exp: 0")
		t.Errorf("Got: %d", r)
		t.Error("Should be no more work running")
	}

	// Shutdown the system with no work.
	if err := w.Shutdown(context.Background()); err != nil {
		t.Fatalf("Should be able to shutdown work cleanly : %s", err)
	}
}

func Test_CancelWorker(t *testing.T) {
	// Create a WaitGroup to know when all 4 jobs have been started.
	var wg sync.WaitGroup
	wg.Add(4)

	// Define a work function that waits to be canceled.
	work := func(ctx context.Context) {
		wg.Done()
		t.Logf("Goroutine running")
		<-ctx.Done()
		t.Logf("Goroutine terminating")
	}

	// Create a worker and start all 4 jobs.
	w, err := worker.New(4)
	if err != nil {
		t.Fatalf("Should be able to create a worker with max 4 : %s", err)
	}

	for i := 0; i < 4; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := w.Start(ctx, work); err != nil {
			t.Fatalf("Should be able to execute work : %s", err)
		}
	}

	// Wait for all 4 jobs to report they are running.
	wg.Wait()

	// Give all the jobs 1 second to shut down cleanly.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := w.Shutdown(ctx); err != nil {
		t.Fatalf("Should be able to shutdown work cleanly : %s", err)
	}
}

func Test_StopWorker(t *testing.T) {
	// Create a WaitGroup to know when all 4 jobs have been started.
	var wg sync.WaitGroup
	wg.Add(4)

	// Define a work function that waits to be canceled.
	work := func(ctx context.Context) {
		wg.Done()
		t.Logf("Goroutine running")
		<-ctx.Done()
		t.Logf("Goroutine terminating")
	}

	var works []string

	// Create a worker and start all 4 jobs.
	w, err := worker.New(4)
	if err != nil {
		t.Fatalf("Should be able to create a worker with max 4 : %s", err)
	}

	for i := 0; i < 4; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		work, err := w.Start(ctx, work)
		if err != nil {
			t.Fatalf("Should be able to execute work : %s", err)
		}
		works = append(works, work)
	}

	// Wait for all 4 jobs to report they are running.
	wg.Wait()

	// Call Stop on all the jobs.
	for _, work := range works {
		w.Stop(work)
	}

	// Wait for all the jobs to finish.
	for i := 0; i < 3; i++ {
		if w.Running() == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Check that the running list is empty.
	if r := w.Running(); r != 0 {
		t.Errorf("Exp: 0")
		t.Errorf("Got: %d", r)
		t.Error("Should be no more work running")
	}

	// Shutdown the system with no work.
	if err := w.Shutdown(context.Background()); err != nil {
		t.Fatalf("Should be able to shutdown work cleanly : %s", err)
	}
}
