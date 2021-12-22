// Package stream provides a set of functions for working with channels of
// generic types. This library uses the new Generics implemention for Go to
// simplify regular patterns of working with channels.
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

// FanIn accepts incoming data channels and forwards returns a single channel
// that receives all the data from the supplied channels.
//
// NOTE: The transfer takes place in a goroutine for each channel
// so ensuring that the context is cancelled or the incoming channels
// are closed is important to ensure that the goroutine is terminated.
func FanIn[T any](ctx context.Context, in ...<-chan T) <-chan T {
	out := make(chan T)

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
