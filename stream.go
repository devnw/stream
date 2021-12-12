// Package stream provides a set of functions for working with channels of
// generic types. This library uses the new Generics implemention for Go to
// simplify regular patterns of working with channels.
package stream

import (
	"context"
)

// Forward accepts an incoming data channel and forwards it to the supplied
// outgoing data channel.
func Forward[T any](ctx context.Context, in <-chan T, out chan<- T) {
	FanOut(ctx, in, out)
}

// FanOut accepts an incoming data channel and forwards the data from it to
// the supplied outgoing data channels.
func FanOut[T any](ctx context.Context, in <-chan T, out ...chan<- T) {
	defer func() {
		defer recover() // catch closed channel errors

		for _, c := range out {
			close(c)
		}
	}()
		
	for {
		select {
		case <- ctx.Done():
			return
		case v, ok := <- in:
			if !ok {
				return
			}

			for _, o := range out {
				select {
				case <- ctx.Done():
					return
				case o <- v:
				}
			}
		}

	}
}

// FanIn accepts incoming data channels and forwards returns a single channel
// that receives all the data from the supplied channels.
func FanIn[T any](ctx context.Context, in ...<-chan T) <-chan T {
	out := make(chan T)

	for _, i := range in {
		go Forward(ctx, i, out)
	}

	return out
}