package worker_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ardanlabs/service/foundation/worker"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

var traceID = "00000000-0000-0000-0000-000000000000"

func TestWorker(t *testing.T) {
	t.Log("Given the need to start work and wait for it to complete.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling multiple jobs", testID)
		{
			// Define a work function that waits to be canceled.
			work := func(ctx context.Context, workKey string, data interface{}) {
				t := data.(*testing.T)
				t.Logf("\t\t%s\tTest %d:\tGoroutine running.", success, testID)
				<-ctx.Done()
				t.Logf("\t\t%s\tTest %d:\tGoroutine terminating.", success, testID)
			}

			// Load 4 jobs in the system.
			var jobs = map[string]worker.JobFunc{
				"script1": work,
				"script2": work,
				"script3": work,
				"script4": work,
			}

			// Create a worker and start all 4 jobs.
			w := worker.New(jobs)
			for key := range jobs {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				if _, err := w.Start(ctx, traceID, key, t); err != nil {
					t.Fatalf("\t\t%s\tTest %d:\tShould be able to execute work: %v", failed, testID, err)
				}
				t.Logf("\t\t%s\tTest %d:\tShould be able to execute work.", success, testID)
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
				t.Errorf("\t\t%s\tTest %d:\tExp: 0", failed, testID)
				t.Errorf("\t\t%s\tTest %d:\tGot: %d", failed, testID, r)
				t.Errorf("\t\t%s\tTest %d:\tShould be no more work running.", failed, testID)
			} else {
				t.Logf("\t\t%s\tTest %d:\tShould be no more work running.", success, testID)
			}

			// Shutdown the system with no work.
			if err := w.Shutdown(context.Background()); err != nil {
				t.Fatalf("\t\t%s\tTest %d:\tShould be able to shutdown work cleanly: %v", failed, testID, err)
			}
			t.Logf("\t\t%s\tTest %d:\tShould be able to shutdown work cleanly.", success, testID)
		}
	}
}

func TestCancelWorker(t *testing.T) {
	t.Log("Given the need to start work and cancel it on shutdown.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling multiple jobs", testID)
		{
			// Create a WaitGroup to know when all 4 jobs have been started.
			var wg sync.WaitGroup
			wg.Add(4)

			// Define a work function that waits to be canceled.
			work := func(ctx context.Context, workKey string, data interface{}) {
				wg.Done()
				t := data.(*testing.T)
				t.Logf("\t\t%s\tTest %d:\tGoroutine running.", success, testID)
				<-ctx.Done()
				t.Logf("\t\t%s\tTest %d:\tGoroutine terminating.", success, testID)
			}

			// Load 4 jobs in the system.
			var jobs = map[string]worker.JobFunc{
				"script1": work,
				"script2": work,
				"script3": work,
				"script4": work,
			}

			// Create a worker and start all 4 jobs.
			w := worker.New(jobs)
			for key := range jobs {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if _, err := w.Start(ctx, traceID, key, t); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to execute work: %v", failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to execute work.", success, testID)
			}

			// Wait for all 4 jobs to report they are running.
			wg.Wait()

			// Give all the jobs 1 second to shutdown cleanly.
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := w.Shutdown(ctx); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to shutdown work cleanly: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to shutdown work cleanly.", success, testID)
		}
	}
}

func TestStopWorker(t *testing.T) {
	t.Log("Given the need to start work and stop work.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling multiple jobs", testID)
		{
			// Create a WaitGroup to know when all 4 jobs have been started.
			var wg sync.WaitGroup
			wg.Add(4)

			// Define a work function that waits to be canceled.
			work := func(ctx context.Context, workKey string, data interface{}) {
				wg.Done()
				t := data.(*testing.T)
				t.Logf("\t%s\tTest %d:\tGoroutine running.", success, testID)
				<-ctx.Done()
				t.Logf("\t%s\tTest %d:\tGoroutine terminating.", success, testID)
			}

			// Load 4 jobs in the system.
			var jobs = map[string]worker.JobFunc{
				"script1": work,
				"script2": work,
				"script3": work,
				"script4": work,
			}

			var works []string

			// Create a worker and start all 4 jobs.
			w := worker.New(jobs)
			for key := range jobs {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				work, err := w.Start(ctx, traceID, key, t)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to execute work: %v", failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to execute work.", success, testID)
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
				t.Errorf("\t\t%s\tTest %d:\tExp: 0", failed, testID)
				t.Errorf("\t\t%s\tTest %d:\tGot: %d", failed, testID, r)
				t.Errorf("\t%s\tTest %d:\tShould be no more work running.", failed, testID)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be no more work running.", success, testID)
			}

			// Shutdown the system with no work.
			if err := w.Shutdown(context.Background()); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to shutdown work cleanly: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to shutdown work cleanly.", success, testID)

		}
	}
}
