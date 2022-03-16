package stream

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type float interface {
	~float32 | ~float64
}

// Initialize the random number generator.
func init() { rand.Seed(time.Now().Unix()) }

func Tst[U ~[]T, T any](
	t *testing.T,
	name string,
	data []U,
	f func(t *testing.T, test []T),
) {
	for _, test := range data {
		testname := fmt.Sprintf("%s-tests[%v]-values[%v]", name, len(data), len(test))
		t.Run(testname, func(t *testing.T) {
			f(t, test)
		})
	}
}

func Int[T integer]() T {
	value := rand.Int()
	return T(value)
}

func Float[T float]() T {
	return T(rand.Float64())
}

func Ints[T integer](size int) []T {
	out := make([]T, size)

	for i := range out {
		out[i] = Int[T]()
	}

	return out
}

func Floats[T float](size int) []T {
	out := make([]T, size)

	for i := range out {
		out[i] = Float[T]()
	}

	return out
}

func IntTests[T integer](tests, cap int) [][]T {
	out := make([][]T, tests)

	for i := range out {
		out[i] = Ints[T](cap)
	}

	return out
}

func FloatTests[T float](tests, cap int) [][]T {
	out := make([][]T, tests)

	for i := range out {
		out[i] = Floats[T](cap)
	}

	return out
}
