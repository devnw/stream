package stream

import (
	"context"
	"testing"

	"go.structs.dev/gen"
)

func Benchmark_Pipe(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1, c2 := make(chan int), make(chan int)
	value := 10

	go Pipe(ctx, c1, c2)

	for n := 0; n < b.N; n++ {
		select {
		case <-ctx.Done():
			b.Fatal("context canceled")
		case c1 <- value:
		case out, ok := <-c2:
			if !ok {
				b.Fatal("c2 closed prematurely")
			}

			if out != value {
				b.Errorf("expected %v, got %v", value, out)
			}
		}
	}
}

func Benchmark_Intercept(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	value := 10

	out := Intercept(ctx, in, func(_ context.Context, v int) (int, bool) {
		return v % 3, true
	})

	for n := 0; n < b.N; n++ {
		select {
		case <-ctx.Done():
			b.Fatal("context canceled")
		case in <- value:
		case out, ok := <-out:
			if !ok {
				b.Fatal("c2 closed prematurely")
			}

			if out != value%3 {
				b.Errorf("expected %v, got %v", value, out)
			}
		}
	}
}

func Benchmark_FanIn(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1, c2 := make(chan int), make(chan int)
	out := FanIn(ctx, c1, c2)

	for n := 0; n < b.N; n++ {
		c1 <- 1
		c2 <- 2

		for i := 0; i < 2; i++ {
			select {
			case <-ctx.Done():
				b.Fatal("context canceled")
			case _, ok := <-out:
				if !ok {
					b.Fatal("out closed prematurely")
				}
			}
		}
	}
}

func Benchmark_FanOut(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out1, out2 := make(chan int), make(chan int), make(chan int)

	go FanOut(ctx, in, out1, out2)

	for n := 0; n < b.N; n++ {
		in <- 1
		<-out1
		<-out2
	}
}

func Benchmark_Distribute(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out1, out2 := make(chan int), make(chan int), make(chan int)

	go Distribute(ctx, in, out1, out2)

	for n := 0; n < b.N; n++ {
		in <- 1

		select {
		case <-out1:
		case <-out2:
		}
	}
}

func Benchmark_Scaler(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testdata := gen.Slice[int](Ints[int](100))

	s := Scaler[int, int]{
		Fn: func(_ context.Context, in int) (int, bool) {
			return in, true
		},
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		// Test that the scaler can be used with a nil context.
		//nolint:staticcheck
		out, err := s.Exec(nil, testdata.Chan(ctx))
		if err != nil {
			b.Errorf("expected no error, got %v", err)
		}

		seen := 0

	tloop:
		for {
			select {
			case <-ctx.Done():
				b.Fatal("context closed")
			case _, ok := <-out:
				if !ok {
					break tloop
				}
				seen++
			}
		}

		if seen != len(testdata) {
			b.Errorf("expected %v, got %v", len(testdata), seen)
		}
	}
}
