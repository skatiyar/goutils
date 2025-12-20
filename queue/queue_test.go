package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/skatiyar/goutils/queue"
)

func TestNewQueue(t *testing.T) {
	cfg := queue.Config{Size: -1, Concurrency: -1, DefaultTimeout: -1}
	q := queue.New(cfg, func(ctx context.Context, v int) (int, error) {
		return v * 2, nil
	})
	if q == nil {
		t.Fatal("expected non-nil queue")
	}
	if q.Running() != 0 {
		t.Fatalf("expected running 0, got %d", q.Running())
	}
	if q.Queued() != 0 {
		t.Fatalf("expected queued 0, got %d", q.Queued())
	}
	if q.Status() != queue.StatusIdle {
		t.Fatalf("expected status idle, got %v", q.Status())
	}
	actualCfg := q.Config()
	if actualCfg.Size != 100 {
		t.Fatalf("expected default size 100, got %d", actualCfg.Size)
	}
	if actualCfg.Concurrency != 10 {
		t.Fatalf("expected default concurrency 10, got %d", actualCfg.Concurrency)
	}
	if actualCfg.DefaultTimeout != queue.DefaultTimeout {
		t.Fatalf("expected default timeout > 0, got %v", actualCfg.DefaultTimeout)
	}
}

func TestPushAndProcess(t *testing.T) {
	cfg := queue.Config{Size: 10, Concurrency: 2, DefaultTimeout: time.Second}
	q := queue.New(cfg, func(ctx context.Context, v int) (int, error) {
		return v * 2, nil
	})

	res := q.Push(context.Background(), 3)
	got, err := res.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 6 {
		t.Fatalf("expected 6, got %v", got)
	}
	if q.Running() != 0 {
		t.Fatalf("expected running 0 after completion, got %d", q.Running())
	}
	if q.Queued() != 0 {
		t.Fatalf("expected queued 0 after completion, got %d", q.Queued())
	}
	if q.Status() != queue.StatusIdle {
		t.Fatalf("expected status idle, got %v", q.Status())
	}
}

func TestRunningAndStatus(t *testing.T) {
	done := make(chan struct{})
	cfg := queue.Config{Size: 10, Concurrency: 2, DefaultTimeout: time.Second}
	q := queue.New(cfg, func(ctx context.Context, v int) (int, error) {
		// block until allowed to finish to keep tasks "running"
		<-done
		return v * 2, nil
	})

	// push three tasks; with concurrency=2 two should be "running"
	r1 := q.Push(context.Background(), 1)
	r2 := q.Push(context.Background(), 2)
	r3 := q.Push(context.Background(), 3)

	// give a moment for internal counters to update
	time.Sleep(1 * time.Millisecond)

	if q.Running() != 2 {
		t.Fatalf("expected running 2, got %d", q.Running())
	}
	if q.Status() != queue.StatusRunning {
		t.Fatalf("expected status running, got %v", q.Status())
	}

	// allow workers to finish
	close(done)

	v1, err1 := r1.Await()
	if err1 != nil || v1 != 2 {
		t.Fatalf("unexpected result r1: %v, %v", v1, err1)
	}
	v2, err2 := r2.Await()
	if err2 != nil || v2 != 4 {
		t.Fatalf("unexpected result r2: %v, %v", v2, err2)
	}
	v3, err3 := r3.Await()
	if err3 != nil || v3 != 6 {
		t.Fatalf("unexpected result r3: %v, %v", v3, err3)
	}

	// give a moment for internal counters to update
	time.Sleep(10 * time.Millisecond)

	if q.Running() != 0 {
		t.Fatalf("expected running 0 after completion, got %d", q.Running())
	}
	if q.Status() != queue.StatusIdle {
		t.Fatalf("expected status idle after completion, got %v", q.Status())
	}
}

func TestShutdownClosesQueue(t *testing.T) {
	cfg := queue.Config{Size: 5, Concurrency: 1, DefaultTimeout: time.Second}
	q := queue.New(cfg, func(ctx context.Context, v int) (int, error) {
		return v, nil
	})

	// shutdown should mark queue closed
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
	if q.Status() != queue.StatusClosed {
		t.Fatalf("expected status closed after shutdown, got %v", q.Status())
	}

	// pushing after shutdown should return ErrQueueClosed
	res := q.Push(context.Background(), 42)
	v, err := res.Await()
	if err != queue.ErrQueueClosed {
		t.Fatalf("expected ErrQueueClosed, got err=%v val=%v", err, v)
	}
}
