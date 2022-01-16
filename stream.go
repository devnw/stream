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
	"crypto/rand"
	"math/big"
)

// Pipe accepts an incoming data channel and pipes it to the supplied
// outgoing data channel.
//
// NOTE: Execute the Pipe function in a goroutine if parallel execution is
// desired. Cancelling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func Pipe[T any](ctx context.Context, in <-chan T, out chan<- T) {
	ctx, _ = _ctx(ctx)

	// Pipe is just a fan-out of a single channel.
	FanOut(ctx, in, out)
}

type InterceptFunc[T, U any] func(context.Context, T) (U, bool)

// Intercept accepts an incoming data channel and a function literal that
// accepts the incoming data and returns data of the same type and a boolean
// indicating whether the data should be forwarded to the output channel.
// The function is executed for each data item in the incoming channel as long
// as the context is not cancelled or the incoming channel remains open.
func Intercept[T, U any](
	ctx context.Context,
	in <-chan T,
	fn InterceptFunc[T, U],
) <-chan U {
	ctx, _ = _ctx(ctx)
	out := make(chan U)

	go func() {
		defer recover()
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
					defer recover()

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
// so ensuring that the context is cancelled or the incoming channels
// are closed is important to ensure that the goroutine is terminated.
func FanIn[T any](ctx context.Context, in ...<-chan T) <-chan T {
	ctx, _ = _ctx(ctx)
	out := make(chan T)

	if len(in) == 0 {
		defer close(out)
		return out
	}

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

// FanOut accepts an incoming data channel and copies the data to each of the
// supplied outgoing data channels.
//
// NOTE: Execute the FanOut function in a goroutine if parallel execution is
// desired. Cancelling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func FanOut[T any](ctx context.Context, in <-chan T, out ...chan<- T) {
	ctx, _ = _ctx(ctx)

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

			for _, o := range out {
				// Closure to catch panic on closed channel write.
				// Continue Loop
				func() {
					defer recover()
					select {
					case <-ctx.Done():
						return
					case o <- v:
					}
				}()
			}
		}

	}
}

// Distribute accepts an incoming data channel and distributes the data among
// the supplied outgoing data channels. This distribution is done stochastically
// using the cryptographic random number generator.
//
// NOTE: Execute the Distribute function in a goroutine if parallel execution is
// desired. Cancelling the context or closing the incoming channel is important
// to ensure that the goroutine is properly terminated.
func Distribute[T any](ctx context.Context, in <-chan T, out ...chan<- T) {
	ctx, _ = _ctx(ctx)

	if len(out) == 0 {
		return
	}

	for {
		// Generate a random number with good entropy to determine which
		// channel the data should be sent to.
		// TODO: Determine if this hinders performance too much and if so,
		// switch to something like fastrand which is used in the runtime
		// for the select statement.
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(len(out))))
		index := int(r.Int64()) % len(out)

		select {
		case <-ctx.Done():
			return
		case v, ok := <-in:
			if !ok {
				return
			}

			// Closure to catch panic on closed channel write.
			func() {
				defer recover()

				select {
				case <-ctx.Done():
					return
				case out[index] <- v:
				}
			}()
		}

	}
}
