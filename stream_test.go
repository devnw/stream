package stream

import (
	"context"
	"testing"
	"time"

	. "go.structs.dev/gen"
)

func PipeTest[U ~[]T, T comparable](
	t *testing.T,
	name string,
	data []U,
) {
	Tst(
		t,
		name,
		data,
		func(t *testing.T, data []T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			c1, c2 := make(chan T), make(chan T)

			go Pipe(ctx, c1, c2)

			go func() {
				for _, v := range data {
					select {
					case <-ctx.Done():
						return
					case c1 <- v:
					}
				}
			}()

			for i := 0; i < len(data); i++ {
				select {
				case <-ctx.Done():
					t.Error("context cancelled")
					return
				case out, ok := <-c2:
					if !ok {
						if i != len(data)-1 {
							t.Fatal("c2 closed prematurely")
						}
						return
					}

					if out != data[i] {
						t.Errorf("expected %v, got %v", data[i], out)
					}
				}
			}
		})
}

func Test_Pipe(t *testing.T) {
	PipeTest(t, "int8", IntTests[int8](100, 1000))
	PipeTest(t, "uint8", IntTests[uint8](100, 1000))
	PipeTest(t, "uint8", IntTests[uint8](100, 1000))
	PipeTest(t, "uint16", IntTests[uint16](100, 1000))
	PipeTest(t, "int32", IntTests[int32](100, 1000))
	PipeTest(t, "uint32", IntTests[uint32](100, 1000))
	PipeTest(t, "int64", IntTests[int64](100, 1000))
	PipeTest(t, "uint64", IntTests[uint64](100, 1000))
	PipeTest(t, "float32", FloatTests[float32](100, 1000))
	PipeTest(t, "float64", FloatTests[float64](100, 1000))
}

func FanInTest[U ~[]T, T comparable](
	t *testing.T,
	name string,
	data []U,
) {
	Tst(
		t,
		name,
		data,
		func(t *testing.T, data []T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			divisor := 5

			if len(data)%divisor != 0 {
				t.Fatalf("data length must be divisible by %v", divisor)
			}

			out := make([]chan T, divisor)

			// Initialize channels
			for i := range out {
				out[i] = make(chan T)
			}

			fan := FanIn(ctx, ReadOnly(out...)...)

			ichan := 0
			cursor := 0
			for i := len(data) / divisor; i <= len(data); i += len(data) / divisor {
				go func(out chan<- T, data []T) {
					defer close(out)

					for _, v := range data {
						select {
						case <-ctx.Done():
							return
						case out <- v:
						}
					}
				}(out[ichan], data[cursor:i])

				cursor = i
				ichan++
			}

			returned := make([]T, len(data))
			for i := 0; i < len(data); i++ {
				select {
				case <-ctx.Done():
					t.Error("context cancelled")
					return
				case out, ok := <-fan:
					if !ok {
						if i != len(data)-1 {
							t.Fatal("c2 closed prematurely")
						}
					}

					returned[i] = out
				}
			}

			diff := Diff(data, returned)
			if len(diff) != 0 {
				t.Errorf("unexpected diff: %v", diff)
			}
		})
}

func Test_FanIn(t *testing.T) {
	FanInTest(t, "int8", IntTests[int8](100, 1000))
	FanInTest(t, "uint8", IntTests[uint8](100, 1000))
	FanInTest(t, "uint8", IntTests[uint8](100, 1000))
	FanInTest(t, "uint16", IntTests[uint16](100, 1000))
	FanInTest(t, "int32", IntTests[int32](100, 1000))
	FanInTest(t, "uint32", IntTests[uint32](100, 1000))
	FanInTest(t, "int64", IntTests[int64](100, 1000))
	FanInTest(t, "uint64", IntTests[uint64](100, 1000))
	FanInTest(t, "float32", FloatTests[float32](100, 1000))
	FanInTest(t, "float64", FloatTests[float64](100, 1000))
}

func InterceptTest[U ~[]T, T signed](
	t *testing.T,
	name string,
	data []U,
) {
	Tst(
		t,
		name,
		data,
		func(t *testing.T, data []T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			in := make(chan T)
			defer close(in)

			out := Intercept(ctx, in, func(_ context.Context, in T) (T, bool) {
				return in % 3, true
			})

			go func() {
				for _, v := range data {
					select {
					case <-ctx.Done():
						return
					case in <- v:
					}
				}
			}()

			for i := 0; i < len(data); i++ {
				select {
				case <-ctx.Done():
					t.Error("context cancelled")
					return
				case out, ok := <-out:
					if !ok {
						if i != len(data)-1 {
							t.Fatal("c2 closed prematurely")
						}
					}

					if out != data[i]%3 {
						t.Errorf("expected %v, got %v", data[i], out)
					}
				}
			}
		})
}
func Test_Intercept(t *testing.T) {
	InterceptTest(t, "int8", IntTests[int8](100, 1000))
	InterceptTest(t, "int8", IntTests[int8](100, 1000))
	InterceptTest(t, "int32", IntTests[int32](100, 1000))
	InterceptTest(t, "int64", IntTests[int64](100, 1000))
}

func Test_Intercept_ChangeType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	integers := Ints[int](100)
	booleans := make([]bool, len(integers))

	for i, v := range integers {
		booleans[i] = v%2 == 0
	}

	out := Intercept(
		ctx,
		Slice[int](integers).Chan(ctx),
		func(_ context.Context, in int) (bool, bool) {
			return in%2 == 0, true
		})

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return
		case out, ok := <-out:
			if !ok {
				return
			}

			if out != booleans[i] {
				t.Errorf("expected %v, got %v", booleans[0], out)
			}
		}
	}
}

func Test_Intercept_NotOk(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	defer close(in)

	go func() {
		in <- 1
		in <- 2
		in <- 3
		in <- 0
	}()

	out := Intercept(ctx, in, func(_ context.Context, v int) (int, bool) {
		if v == 0 {
			return 0, true
		}

		return v, false
	})

	select {
	case <-ctx.Done():
		t.Fatal("context cancelled")
	case out, ok := <-out:
		if !ok {
			t.Fatal("in closed prematurely")
		}

		if out != 0 {
			t.Errorf("expected 0, got %v", out)
		}
	}
}

func Test_Intercept_ClosedChan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)

	out := Intercept(
		ctx,
		in,
		func(_ context.Context, v int) (int, bool) {
			return v, false
		})

	close(in)

	<-out
}

func Test_Intercept_Cancelled_On_Wait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	defer close(in)

	// Setup intercept
	out := Intercept(ctx, in, func(_ context.Context, v int) (int, bool) {
		return v, true
	})

	// Push to in
	in <- 1

	// Wait for intercept routine to be scheduled
	time.Sleep(time.Millisecond)

	// Cancel the routine
	cancel()

	_, ok := <-out
	if ok {
		t.Error("expected closed channel")
	}
}

func Test_FanOut_Cancelled_On_Wait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := make(chan int), make(chan int)
	defer close(in)
	defer close(out)

	go func() {
		defer cancel()
		in <- 1
	}()

	FanOut(ctx, in, out)
}

func DistributeTest[U ~[]T, T comparable](
	t *testing.T,
	name string,
	data []U,
) {
	Tst(
		t,
		name,
		data,
		func(t *testing.T, data []T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			c1, c2, c3 := make(chan T), make(chan T), make(chan T)

			go Distribute(ctx, Slice[T](data).Chan(ctx), c1, c2, c3)

			c1total, c2total, c3total := 0, 0, 0
			for i := 0; i < len(data); i++ {
				select {
				case <-ctx.Done():
					t.Error("context cancelled")
					return
				case out, ok := <-c1:
					if !ok {
						return
					}

					if out != data[i] {
						t.Errorf("expected %v, got %v", data[i], out)
					}
					c1total++
				case out, ok := <-c2:
					if !ok {
						return
					}

					if out != data[i] {
						t.Errorf("expected %v, got %v", data[i], out)
					}
					c2total++
				case out, ok := <-c3:
					if !ok {
						return
					}

					if out != data[i] {
						t.Errorf("expected %v, got %v", data[i], out)
					}
					c3total++
				}
			}

			t.Logf("c1: %v", c1total)
			t.Logf("c2: %v", c2total)
			t.Logf("c3: %v", c3total)

			ctotal := c1total + c2total + c3total
			if ctotal != len(data) {
				t.Errorf("expected %v, got %v", len(data), ctotal)
			}
		})
}

func Test_Distribute(t *testing.T) {
	DistributeTest(t, "int8", IntTests[int8](100, 1000))
	DistributeTest(t, "uint8", IntTests[uint8](100, 1000))
	DistributeTest(t, "uint8", IntTests[uint8](100, 1000))
	DistributeTest(t, "uint16", IntTests[uint16](100, 1000))
	DistributeTest(t, "int32", IntTests[int32](100, 1000))
	DistributeTest(t, "uint32", IntTests[uint32](100, 1000))
	DistributeTest(t, "int64", IntTests[int64](100, 1000))
	DistributeTest(t, "uint64", IntTests[uint64](100, 1000))
	DistributeTest(t, "float32", FloatTests[float32](100, 1000))
	DistributeTest(t, "float64", FloatTests[float64](100, 1000))
}

func Test_Distribute_Cancelled_On_Wait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := make(chan int), make(chan int)
	defer close(in)
	defer close(out)

	go func() {
		defer cancel()
		in <- 1
	}()

	Distribute(ctx, in, out)
}

func Test_Distribute_ZeroOut(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	defer close(in)

	Distribute(ctx, in)
}

func Test_FanOut(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1, c2, c3 := make(chan int), make(chan int), make(chan int)
	var c4 chan int
	data := Ints[int](1000)

	go FanOut(ctx, Slice[int](data).Chan(ctx), c1, c2, c3, c4)

	seen := make(map[int]int)
	for i := 0; i < len(data)*3; i++ {
		select {
		case <-ctx.Done():
			t.Fatal("context cancelled")
			return
		case _, ok := <-c1:
			if !ok {
				return
			}

			seen[1]++
		case _, ok := <-c2:
			if !ok {
				return
			}

			seen[2]++
		case _, ok := <-c3:
			if !ok {
				return
			}

			seen[3]++
		case _, ok := <-c4:
			if !ok {
				return
			}

			seen[4]++
		}
	}

	if len(seen) != 3 {
		t.Fatalf("expected %v, got %v", len(data)-1, len(seen))
	}

	for k, v := range seen {
		if k == 4 {
			if v > 0 {
				t.Fatalf("expected %v, got %v", 0, v)
			}
		}

		if v != len(data) {
			t.Fatalf("Chan C%v: expected %v, got %v", k, len(data), v)
		}
	}
}

func Test_FanOut_ZeroOut(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	defer close(in)

	FanOut(ctx, in)
}

func Test_FanIn_ZeroIn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	defer close(in)

	FanIn[int](ctx)
}
