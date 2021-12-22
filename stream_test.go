package stream

import (
	"fmt"
	"testing"
)

func PipeTest[U ~[]T, T any](
	t *testing.T,
	name string,
	data map[string][]U,
) {
	Tst(
		t,
		name,
		data,
		func(t *testing.T, testdata []U) {
			for _, v := range testdata {

				for _, v2 := range v {
					fmt.Printf("%v-%v", v, v2)
				}
			}
		})
}

func Test_Pipe(t *testing.T) {
	PipeTest(
		t,
		"int",
		map[string][][]int{
			"rand1": IntTests[int](100, 1000),
		},
	)
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
