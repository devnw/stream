package stream

import (
	"context"
	"testing"
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
			b.Fatal("context cancelled")
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

	out := Intercept(ctx, in, func(v int) (int, bool) {
		return v % 3, true
	})

	for n := 0; n < b.N; n++ {
		select {
		case <-ctx.Done():
			b.Fatal("context cancelled")
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
				b.Fatal("context cancelled")
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
