// Package stream provides a set of generic functions for working concurrent
// design patterns in Go.
//
// It is recommended to use the package via the following import:
//
//     import . "go.atomizer.io/stream"
//
// Using the `.` import allows for functions to be called directly as if
// the functions were in the same namespace without the need to append
// the package name.
package stream

import (
	"context"
)

// Pipe accepts an incoming data channel and pipes it to the supplied
// outgoing data channel.
//
// NOTE: Execute the Pipe function in a goroutine if parallel execution is
// desired. Cancelling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func Pipe[T any](ctx context.Context, in <-chan T, out chan<- T) {

	// Pipe is just a fan-out of a single channel.
	FanOut(ctx, in, out)
}

// Intercept accepts an incoming data channel and a function literal that
// accepts the incoming data and returns data of the same type and a boolean
// indicating whether the data should be forwarded to the output channel.
// The function is executed for each data item in the incoming channel as long
// as the context is not cancelled or the incoming channel remains open.
func Intercept[T any](ctx context.Context, in <-chan T, fn func(in T) (T, bool)) <-chan T {
	out := make(chan T)

	go func() {
		defer recover() // catch closed channel errors
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
					// TODO: Should something happen with this panic data?
					defer recover() // catch panic

					// Determine if the function was successful
					result, ok := fn(v)
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
// so ensuring that the context is cancelled or the incoming channels
// are closed is important to ensure that the goroutine is terminated.
func FanIn[T any](ctx context.Context, in ...<-chan T) <-chan T {
	out := make(chan T)
	defer func() {
		go func() {
			<-ctx.Done()
			close(out)
		}()
	}()

	for _, i := range in {
		// Pipe the result of the channel to the output channel.
		go Pipe(ctx, i, out)
	}

	return out
}

// FanOut accepts an incoming data channel and forwards the data from it to
// the supplied outgoing data channels.
//
// NOTE: Execute the FanOut function in a goroutine if parallel execution is
// desired. Cancelling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func FanOut[T any](ctx context.Context, in <-chan T, out ...chan<- T) {
	defer func() {
		defer recover() // catch closed channel errors
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-in:
			if !ok {
				return
			}

			// Closure to catch panic on closed channel write.
			func() {
				defer recover() // catch closed channel errors

				for _, o := range out {
					select {
					case <-ctx.Done():
						return
					case o <- v:
					}
				}
			}()
		}

	}
}
