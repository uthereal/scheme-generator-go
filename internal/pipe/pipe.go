package pipe

// Pipe is a generic type wrapper around any type T.
type Pipe[T any] struct {
	value T
}

// NewPipe is a helper to instantiate the pipe with type inference.
func NewPipe[T any](val T) Pipe[T] {
	return Pipe[T]{value: val}
}

// Do takes a function that accepts T and an extra argument, and returns a
// modified T. This allows you to pass standard library functions like
// strings.TrimPrefix directly.
func (p Pipe[T]) Do(
	fn func(a T, b T) T,
	arg T,
) Pipe[T] {
	return Pipe[T]{value: fn(p.value, arg)}
}

// DoUnary handles functions that only operate on the value itself.
func (p Pipe[T]) DoUnary(
	fn func(val T) T,
) Pipe[T] {
	return Pipe[T]{value: fn(p.value)}
}

// Unwrap returns the underlying value of type T.
func (p Pipe[T]) Unwrap() T {
	return p.value
}
