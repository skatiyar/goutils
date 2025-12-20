# Queue Library

A generic, concurrent queue implementation in Go with built-in worker pool semantics.

## Interface

```go
type Queue[T, R any] interface {
    Push(ctx context.Context, value T) (async.Result[R], error)
    Shutdown(ctx context.Context) error
    
    Queued() int
    Running() int
    Status() Status
}

type Status int

const (
    StatusIdle Status = iota
    StatusRunning
    StatusClosed
)

type Config struct {
    Size           int
    Concurrency    int
    DefaultTimeout time.Duration // 0 = block forever, used when ctx has no deadline
}

func New[T, R any](
    cfg Config,
    process func(context.Context, T) (R, error),
) Queue[T, R]

var (
    ErrQueueClosed = errors.New("queue is closed")
    ErrPushTimeout = errors.New("push timeout exceeded")
)
```

## Design Decisions

### Type Parameters

- **T**: Input value type
- **R**: Result type after processing T

The queue processes items internally using the `process` function provided during construction.

### Push Behavior

`Push(ctx context.Context, value T) async.Result[R]`

- Returns immediately with a future/promise (`async.Result[R]`)
- Blocks if queue is full
- Respects context deadline/cancellation
- Falls back to `DefaultTimeout` if context has no deadline
- Returns `ErrQueueClosed` if queue is closed
- Returns `ErrPushTimeout` if timeout exceeded
- If context is cancelled during processing, the `process` function receives the cancelled context and `async.Result[R]` returns context error

**Timeout Priority:**
1. Context deadline (if present)
2. `DefaultTimeout` from config (if context has no deadline)
3. Block forever if `DefaultTimeout` is 0 and context has no deadline

### Shutdown Behavior

`Shutdown(ctx context.Context) error`

- Stops accepting new items
- Waits for all in-flight items to complete processing
- If context times out, kills workers immediately and returns error
- Multiple calls are no-op (no error returned)

### Queue Capacity

- Static size set during construction via `Config.Size`
- Queue blocks on full (with timeout rules above)

### Status

`Status() Status`

- **StatusIdle**: Queue created but no items processing
- **StatusRunning**: Items actively being processed
- **StatusClosed**: Queue shut down

**Note:** Status is purely informational for metrics/debugging. Race conditions exist between checking status and calling methods—don't use for flow control.

### Metrics

- `Queued()`: Number of items waiting to be processed (excludes currently processing items)
- `Running()`: Number of items currently being processed
- `Status()`: Current queue state

### Processing Function

The `process` function receives:
- The same context passed to `Push`
- The input value of type T

If the context is cancelled during processing, it's the `process` function's responsibility to respect cancellation.

## Example Usage

```go
// Create queue that processes integers and returns their squares
cfg := Config{
    Size:           100,
    Concurrency:    5,
    DefaultTimeout: 10 * time.Second,
}

q := New(cfg, func(ctx context.Context, val int) (int, error) {
    // Simulate work
    select {
    case <-time.After(time.Second):
        return val * val, nil
    case <-ctx.Done():
        return 0, ctx.Err()
    }
})

// Push with custom timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := q.Push(ctx, 42)
if err != nil {
    if errors.Is(err, ErrQueueClosed) {
        // Queue was closed
    } else if errors.Is(err, ErrPushTimeout) {
        // Timeout exceeded
    }
    return err
}

// Get result (blocks until processing completes)
value, err := result.Get()

// Graceful shutdown with 30s timeout
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
q.Shutdown(shutdownCtx)
```

## Architecture Notes

This is a **worker pool with queuing**, not just a queue:

- Items are processed internally by the queue
- No `Pop()` method—processing logic is bundled with the queue
- The queue manages worker goroutines based on `Concurrency` setting

If you need a plain queue where the caller controls processing, this interface isn't appropriate. Consider a standard channel or a queue with `Pop()` semantics instead.
