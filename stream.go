// Package stream provides a set of generic functions for working concurrent
// design patterns in Go.
//
// It is recommended to use the package via the following import:
//
//	import . "go.atomizer.io/stream"
//
// Using the `.` import allows for functions to be called directly as if
// the functions were in the same namespace without the need to append
// the package name.
package stream

import (
	"context"
	"reflect"
	"sync"

	"go.structs.dev/gen"
)

type readonly[T any] interface {
	<-chan T | chan T
}

type writeonly[T any] interface {
	chan<- T | chan T
}

// Pipe accepts an incoming data channel and pipes it to the supplied
// outgoing data channel.
//
// NOTE: Execute the Pipe function in a goroutine if parallel execution is
// desired. Canceling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func Pipe[In readonly[T], Out writeonly[T], T any](
	ctx context.Context, in In, out Out,
) {
	ctx = _ctx(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-in:
			if !ok {
				return
			}

			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}
}

type InterceptFunc[T, U any] func(context.Context, T) (U, bool)

// Intercept accepts an incoming data channel and a function literal that
// accepts the incoming data and returns data of the same type and a boolean
// indicating whether the data should be forwarded to the output channel.
// The function is executed for each data item in the incoming channel as long
// as the context is not canceled or the incoming channel remains open.
func Intercept[In readonly[T], T, U any](
	ctx context.Context,
	in In,
	fn InterceptFunc[T, U],
) <-chan U {
	ctx = _ctx(ctx)
	out := make(chan U)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				// Executing this in a function literal ensures that any panic
				// will be caught during execution of the function
				func() {
					// Determine if the function was successful
					result, ok := fn(ctx, v)
					if !ok {
						return
					}

					// Execute the function against the incoming value
					// and send the result to the output channel.
					select {
					case <-ctx.Done():
						return
					case out <- result:
					}
				}()
			}
		}
	}()

	return out
}

// FanIn accepts incoming data channels and forwards returns a single channel
// that receives all the data from the supplied channels.
//
// NOTE: The transfer takes place in a goroutine for each channel
// so ensuring that the context is canceled or the incoming channels
// are closed is important to ensure that the goroutine is terminated.
func FanIn[In readonly[T], T any](ctx context.Context, in ...In) <-chan T {
	ctx = _ctx(ctx)
	out := make(chan T)

	if len(in) == 0 {
		defer close(out)
		return out
	}

	var wg sync.WaitGroup
	defer func() {
		go func() {
			wg.Wait()
			close(out)
		}()
	}()

	wg.Add(len(in))
	for _, i := range in {
		// Pipe the result of the channel to the output channel.
		go func(i <-chan T) {
			defer wg.Done()
			Pipe(ctx, i, out)
		}(i)
	}

	return out
}

// FanOut accepts an incoming data channel and copies the data to each of the
// supplied outgoing data channels.
//
// NOTE: Execute the FanOut function in a goroutine if parallel execution is
// desired. Canceling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func FanOut[In readonly[T], Out writeonly[T], T any](
	ctx context.Context, in In, out ...Out,
) {
	ctx = _ctx(ctx)

	if len(out) == 0 {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-in:
			if !ok {
				return
			}

			// Closure to catch panic on closed channel write.
			selectCases := make([]reflect.SelectCase, 0, len(out)+1)

			// 0 index is context
			selectCases = append(selectCases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done()),
			})

			for _, outc := range out {
				// Skip nil channels until they are non-nil
				if outc == nil {
					continue
				}

				selectCases = append(selectCases, reflect.SelectCase{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(outc),
					Send: reflect.ValueOf(v),
				})
			}

			for len(selectCases) > 1 {
				chosen, _, _ := reflect.Select(selectCases)

				// The context was canceled.
				if chosen == 0 {
					return
				}

				selectCases = gen.Exclude(selectCases, selectCases[chosen])
			}
		}
	}
}

// Distribute accepts an incoming data channel and distributes the data among
// the supplied outgoing data channels using a dynamic select statement.
//
// NOTE: Execute the Distribute function in a goroutine if parallel execution is
// desired. Canceling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func Distribute[In readonly[T], Out writeonly[T], T any](
	ctx context.Context, in In, out ...Out,
) {
	ctx = _ctx(ctx)

	if len(out) == 0 {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-in:
			if !ok {
				return
			}

			selectCases := make([]reflect.SelectCase, 0, len(out)+1)
			for _, outc := range out {
				selectCases = append(selectCases, reflect.SelectCase{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(outc),
					Send: reflect.ValueOf(v),
				})
			}
			selectCases = append(selectCases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done()),
			})
			_, _, _ = reflect.Select(selectCases)
		}
	}
}

// Drain accepts a channel and drains the channel until the channel is closed
// or the context is canceled.
func Drain[U readonly[T], T any](ctx context.Context, in U) {
	ctx = _ctx(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-in:
				if !ok {
					return
				}
			}
		}
	}()
}

// Any accepts an incoming data channel and converts the channel to a readonly
// channel of the `any` type.
func Any[U readonly[T], T any](ctx context.Context, in U) <-chan any {
	return Intercept(ctx, in, func(_ context.Context, in T) (any, bool) {
		return in, true
	})
}
