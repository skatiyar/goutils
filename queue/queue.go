package queue

type Queue[T any] interface {
	Length() int
	Started() bool
	Running() int
	WorkersList() []T
	Idle() bool
	Concurrency() int
}
