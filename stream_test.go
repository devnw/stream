package stream

import (
	"constraints"
	"context"
	"testing"

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

func InterceptTest[U ~[]T, T constraints.Signed](
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

			out := Intercept(ctx, in, func(v T) (T, bool) {
				return v % 3, true
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

// func (test *Test[any]) Run(t *testing.T) {
// 	for name, test := range test.data {
// 		t.Run(name, func(t *testing.T) {
// 			t.Log("------------------------------------------------")
// 			t.Log(test)
// 			t.Log("------------------------------------------------")
// 		})
// 	}
// }

// func Test_Int(t *testing.T) {

// 	// t.Log(IntTests[int8](100, 1000))
// 	// t.Log(IntTests[uint8](100, 1000))
// 	// t.Log(IntTests[int16](100, 1000))
// 	// t.Log(IntTests[uint16](100, 1000))
// 	// t.Log(IntTests[int32](100, 1000))
// 	// t.Log(IntTests[uint32](100, 1000))
// 	// t.Log(IntTests[int64](100, 1000))
// 	// t.Log(IntTests[uint64](100, 1000))
// 	// t.Log(FloatTests[float32](100, 1000))
// 	// t.Log(FloatTests[float64](100, 1000))

// 	tests := Test[int8]{
// 		data: map[string][][]int8{
// 			"t1": IntTests[int8](100, 1000),
// 		},
// 	}

// 	tests.Run(t)
// }

// func Test_Pipe_Single(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Fatal(r)
// 		}
// 	}()

// 	c1, c2 := make(chan int), make(chan int)

// 	go Pipe(ctx, c1, c2)

// 	go func() {
// 		c1 <- 1
// 	}()

// 	out, ok := <-c2
// 	if !ok {
// 		t.Fatal("c2 closed")
// 	}

// 	if out != 1 {
// 		t.Fatalf("expected 1, got %d", out)
// 	}

// 	close(c1)
// 	timer := time.NewTimer(time.Second)

// 	select {
// 	case <-ctx.Done():
// 		t.Fatal("context cancelled")
// 	case <-timer.C:
// 		t.Fatal("timeout")
// 	case <-c2:
// 	}
// }

// type testy struct {
// 	name string
// }

// type testy2 io.ReadCloser

// func c[T any](in ...T) chan T {
// 	return make(chan T)
// }

// func Test_Forward_int(t *testing.T) {
// 	testdata := map[string]struct {
// 		data any
// 	}{
// 		"ints":    []int{1, 2, 3},
// 	}

// 	for name, test := range testdata {
// 		t.Run(name, func(t *testing.T) {
// 		})
// 	}
// }

// func Test_Forward(t *testing.T) {
// 	testdata := map[string]struct {
// 		data any
// 	}{
// 		"ints":    []int{1, 2, 3},
// 		"int32":   []int32{1, 2, 3},
// 		"ints64":  []int64{1, 2, 3},
// 		"bool":    []bool{true, false, true},
// 		"string":  []string{"test1", "test2", "test3"},
// 		"float32": []float32{1.99, 2.99, 3.99},
// 		"float64": []float32{1.99, 2.99, 3.99},
// 		"testy": []testy{
// 			{"test1"},
// 			{"test2"},
// 			{"test3"},
// 		},
// 	}

// 	for name, test := range testdata {
// 		t.Run(name, func(t *testing.T) {
// 			// ctx, cancel := context.WithCancel(context.Background())
// 			// defer cancel()

// 			// defer func() {
// 			// 	if r := recover(); r != nil {
// 			// 		t.Fatal(r)
// 			// 	}
// 			// }()

// 			// c1, c2 := make(chan int), make(chan int)

// 			// go Pipe(ctx, c1, c2)

// 			// go func() {
// 			// 	c1 <- 1
// 			// }()

// 			// out, ok := <-c2
// 			// if !ok {
// 			// 	t.Fatal("c2 closed")
// 			// }

// 			// if out != 1 {
// 			// 	t.Fatalf("expected 1, got %d", out)
// 			// }

// 			// close(c1)
// 			// timer := time.NewTimer(time.Second)

// 			// select {
// 			// case <-ctx.Done():
// 			// 	t.Fatal("context cancelled")
// 			// case <-timer.C:
// 			// 	t.Fatal("timeout")
// 			// case <-c2:
// 			// }
// 		})
// 	}
// }

// func Test_FanOut(t *testing.T) {
// 	testdata := map[string]struct {
// 	}{
// 		"": {},
// 	}

// 	for name, test := range testdata {
// 		t.Run(name, func(t *testing.T) {

// 		})
// 	}
// }
// func Test_FanIn(t *testing.T) {
// 	testdata := map[string]struct {
// 	}{
// 		"": {},
// 	}

// 	for name, test := range testdata {
// 		t.Run(name, func(t *testing.T) {

// 		})
// 	}
// }
