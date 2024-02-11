package stream

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.devnw.com/gen"
)

var emptyFn = func(context.Context, any) (any, bool) { return 0, true }
var nosendFn = func(context.Context, any) (any, bool) { return 0, false }

func ScalerTest[U ~[]T, T comparable](
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

			testdata := gen.Slice[T](data)

			integers := testdata.Map()

			s := Scaler[T, T]{
				Fn: func(_ context.Context, in T) (T, bool) {
					return in, true
				},
			}

			// Test that the scaler can be used with a nil context.
			//nolint:staticcheck // nil context on purpose
			out, err := s.Exec(nil, testdata.Chan(ctx))
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

		tloop:
			for {
				select {
				case <-ctx.Done():
					t.Fatal("context closed")
				case v, ok := <-out:
					if !ok {
						break tloop
					}

					integers[v] = true
				}
			}

			for k, v := range integers {
				seen, ok := v.(bool)
				if !ok {
					t.Errorf("expected bool, got %T", v)
				}

				if !seen {
					t.Errorf("expected %v, got %v for %v", true, v, k)
				}
			}
		})
}

func Test_Scaler_Exec(t *testing.T) {
	ScalerTest(t, "int8", IntTests[int8](10, 100))
	ScalerTest(t, "uint8", IntTests[uint8](10, 100))
	ScalerTest(t, "uint8", IntTests[uint8](10, 100))
	ScalerTest(t, "uint16", IntTests[uint16](10, 100))
	ScalerTest(t, "int32", IntTests[int32](10, 100))
	ScalerTest(t, "uint32", IntTests[uint32](10, 100))
	ScalerTest(t, "int64", IntTests[int64](10, 100))
	ScalerTest(t, "uint64", IntTests[uint64](10, 100))
	ScalerTest(t, "float32", FloatTests[float32](10, 100))
	ScalerTest(t, "float64", FloatTests[float64](10, 100))
}

func Test_Scaler_NilFn(t *testing.T) {
	s := Scaler[any, any]{}

	//nolint:staticcheck // nil context on purpose
	_, err := s.Exec(nil, nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func Test_Scaler_NilCtx(t *testing.T) {
	s := Scaler[any, any]{
		Fn: emptyFn,
	}

	// Overwrite the default context with a cancelable context.
	var cancel context.CancelFunc
	defaultCtx, cancel = context.WithCancel(context.Background())

	// Fix the default context after the test completes
	t.Cleanup(func() {
		defaultCtx = context.Background()
	})

	cancel()

	// Test that the scaler can be used with a nil context.
	//nolint:staticcheck // nil context on purpose
	out, err := s.Exec(nil, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	select {
	case <-time.After(time.Second):
		t.Errorf("expected no timeout, got timeout")
	case _, ok := <-out:
		if ok {
			t.Errorf("expected out to be closed")
		}
	}
}

func Test_Scaler_CloseIn(t *testing.T) {
	s := Scaler[any, any]{
		Fn: emptyFn,
	}

	in := make(chan any)
	close(in)

	// Test that the scaler can be used with a nil context.
	//nolint:staticcheck // nil context on purpose
	out, err := s.Exec(nil, in)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	select {
	case <-time.After(time.Second):
		t.Errorf("expected no timeout, got timeout")
	case _, ok := <-out:
		if ok {
			t.Errorf("expected out to be closed")
		}
	}
}

func Test_Scaler_l2ctx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := Scaler[any, any]{
		Wait: time.Minute,
		Fn:   emptyFn,
	}

	in := make(chan any)

	// Test that the scaler can be used with a nil context.
	out, err := s.Exec(ctx, in)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Trigger the internal loop of the scaler
	in <- 1

	// Cancel the context while it's waiting to
	// scale to layer 2.
	cancel()

	select {
	case <-time.After(time.Second):
		t.Errorf("expected no timeout, got timeout")
	case _, ok := <-out:
		if ok {
			t.Errorf("expected out to be closed")
		}
	}
}

func Test_Scaler_layer2_ctx1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := Scaler[any, any]{
		Wait: time.Minute,
		Life: time.Minute,
		Fn:   emptyFn,
	}

	out := s.layer2(ctx, nil)

	select {
	case <-time.After(time.Second):
		t.Errorf("expected no timeout, got timeout")
	case _, ok := <-out:
		if ok {
			t.Errorf("expected out to be closed")
		}
	}
}

func Test_Scaler_layer2_closeIn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := Scaler[any, any]{
		Wait: time.Minute,
		Life: time.Minute,
		Fn:   emptyFn,
	}

	in := make(chan any)
	close(in)

	out := s.layer2(ctx, in)

	select {
	case <-time.After(time.Second):
		t.Errorf("expected no timeout, got timeout")
	case _, ok := <-out:
		if ok {
			t.Errorf("expected out to be closed")
		}
	}
}

func Test_Scaler_layer2_ctx2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := Scaler[any, any]{
		Wait: time.Minute,
		Life: time.Minute,
		Fn:   emptyFn,
	}

	in := make(chan any)
	defer close(in)

	out := s.layer2(ctx, in)

	// Push data to the channel to trigger the internal loop and block
	in <- 1
	cancel()
	<-ctx.Done()

	for i := 0; i < 1000; i++ {
		select {
		case <-time.After(time.Second):
			t.Errorf("expected no timeout, got timeout")
		case _, ok := <-out:
			if !ok {
				return
			}
		}
	}

	t.Errorf("expected out to be closed")
}

func Test_Scaler_layer2_nosend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := Scaler[any, any]{
		Wait: time.Minute,
		Life: time.Minute,
		Fn:   nosendFn,
	}

	in := make(chan any)
	defer close(in)

	out := s.layer2(ctx, in)

	// Push data to the channel to trigger the internal loop and block
	in <- 1

	select {
	case <-time.After(time.Millisecond):
	case <-out:
		t.Fatalf("expected 0 data to be sent, got 1")
	}
}

func TestTickDur(t *testing.T) {
	testCases := []struct {
		name        string
		tick        DurationScaler
		duration    time.Duration
		currentStep int
		expected    time.Duration
	}{
		{
			name:        "Test case 1",
			tick:        DurationScaler{Interval: 3, ScalingFactor: 0.1, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 3,
			expected:    11 * time.Second,
		},
		{
			name:        "Test case 2",
			tick:        DurationScaler{Interval: 5, ScalingFactor: -0.1, originalDuration: 20 * time.Second},
			duration:    20 * time.Second,
			currentStep: 10,
			expected:    18 * time.Second,
		},
		{
			name:        "Test case 3",
			tick:        DurationScaler{Interval: 2, ScalingFactor: 0.5, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 4,
			expected:    15 * time.Second,
		},
		{
			name:        "Test case 4",
			tick:        DurationScaler{Interval: 4, ScalingFactor: -0.5, originalDuration: 30 * time.Second},
			duration:    30 * time.Second,
			currentStep: 8,
			expected:    15 * time.Second,
		},
		{
			name:        "Test case 5",
			tick:        DurationScaler{Interval: 3, ScalingFactor: 0.1, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 2,
			expected:    10 * time.Second,
		},
		{
			name:        "Test case 6: Step is divisible, modifier in range",
			tick:        DurationScaler{Interval: 3, ScalingFactor: 0.1, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 3,
			expected:    11 * time.Second,
		},
		{
			name:        "Test case 7: Step is not divisible, modifier in range",
			tick:        DurationScaler{Interval: 3, ScalingFactor: 0.1, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 2,
			expected:    10 * time.Second,
		},
		{
			name:        "Test case 8: Step is divisible, modifier is zero",
			tick:        DurationScaler{Interval: 3, ScalingFactor: 0, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 3,
			expected:    10 * time.Second,
		},
		{
			name:        "Test case 9: Step is divisible, modifier is out of range",
			tick:        DurationScaler{Interval: 3, ScalingFactor: 1, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 3,
			expected:    10 * time.Second,
		},
		{
			name:        "Test case 10: Step is zero, modifier in range",
			tick:        DurationScaler{Interval: 0, ScalingFactor: 0.1, originalDuration: 10 * time.Second},
			duration:    10 * time.Second,
			currentStep: 3,
			expected:    10 * time.Second,
		},
		{
			name: "Test case 6: Step number decreases",
			tick: DurationScaler{
				Interval:         2,
				ScalingFactor:    0.5,
				originalDuration: 10 * time.Second,
				lastInterval:     4,
			},
			duration:    15 * time.Second,
			currentStep: 2,
			expected:    10 * time.Second,
		},
		{
			name: "Test case 7: testing below minwait",
			tick: DurationScaler{
				Interval:         1,
				ScalingFactor:    -0.999,
				originalDuration: time.Millisecond * 2,
				lastInterval:     0,
			},
			duration:    MinWait,
			currentStep: 1,
			expected:    MinWait,
		},
		{
			name: "Test case 8: testing below minwait",
			tick: DurationScaler{
				Interval:         1,
				ScalingFactor:    -0.999,
				originalDuration: time.Millisecond * 900,
				lastInterval:     0,
			},
			duration:    MinWait,
			currentStep: 1,
			expected:    MinWait,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := (&tc.tick).scaledDuration(tc.duration, tc.currentStep)
			if result != tc.expected {
				t.Errorf("Expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}

func FuzzTick(f *testing.F) {
	f.Fuzz(func(t *testing.T, step, cStep int, mod float64, orig, dur int64) {
		tick := &DurationScaler{
			Interval:         step,
			ScalingFactor:    mod,
			originalDuration: time.Duration(orig),
		}

		v := tick.scaledDuration(time.Duration(dur), cStep)
		if v < 0 {
			t.Fatalf("negative duration: %v", v)
		}
	})
}

func FuzzScaler(f *testing.F) {
	interceptFunc := func(_ context.Context, t int) (string, bool) {
		return fmt.Sprintf("%d", t), true
	}

	f.Fuzz(func(
		t *testing.T,
		wait, life int64,
		step, _ int,
		mod float64,
		max uint,
		in int,
	) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tick := DurationScaler{
			Interval:      step,
			ScalingFactor: mod,
		}

		// Initialize Scaler
		scaler := Scaler[int, string]{
			Wait:         time.Millisecond * time.Duration(wait),
			Life:         time.Millisecond * time.Duration(life),
			Fn:           interceptFunc,
			WaitModifier: tick,
			Max:          max,
		}

		// Create a simple input channel
		input := make(chan int, 1)
		defer close(input)

		// Execute the Scaler
		out, err := scaler.Exec(ctx, input)
		if err != nil {
			t.Errorf("Scaler Exec failed: %v", err)
			t.Fail()
		}

		// Send input value and check output
		input <- in

		select {
		case <-ctx.Done():
			t.Errorf("Scaler Exec timed out")
			t.Fail()
		case res := <-out:
			if res != fmt.Sprintf("%d", in) {
				t.Errorf("Scaler Exec failed: expected %d, got %s", in, res)
				t.Fail()
			}

			t.Logf("Scaler Exec succeeded: expected %d, got %s", in, res)
		}
	})
}

//nolint:gocognit // This is a test function
func Test_Scaler_Max(t *testing.T) {
	tests := map[string]struct {
		max      uint
		send     int
		expected int
	}{
		"max 0": {
			max:      0,
			send:     1000,
			expected: 1000,
		},
		"max 1": {
			max:      1,
			send:     10,
			expected: 10,
		},
		"max 2": {
			max:      2,
			send:     10,
			expected: 10,
		},
		"max 3": {
			max:      3,
			send:     10,
			expected: 10,
		},
		"max 4": {
			max:      4,
			send:     100,
			expected: 100,
		},
		"max 1000": {
			max:      1000,
			send:     10000,
			expected: 10000,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			inited := 0
			initedMu := sync.Mutex{}
			release := make(chan struct{})

			interceptFunc := func(_ context.Context, t int) (int, bool) {
				defer func() {
					initedMu.Lock()
					defer initedMu.Unlock()
					inited--
				}()

				initedMu.Lock()
				inited++
				initedMu.Unlock()

				<-release

				return t, true
			}

			// Initialize Scaler
			scaler := Scaler[int, int]{
				Wait: time.Millisecond,
				Life: time.Millisecond,
				Fn:   interceptFunc,
				Max:  test.max,
			}

			// Create a simple input channel
			input := make(chan int, test.send)
			defer close(input)

			for i := 0; i < test.send; i++ {
				input <- i
			}

			// Execute the Scaler
			out, err := scaler.Exec(ctx, input)
			if err != nil {
				t.Errorf("Scaler Exec failed: %v", err)
				t.Fail()
			}

			recv := 1

		tloop:
			for {
				select {
				case <-ctx.Done():
					t.Errorf("Scaler Exec timed out")
				case _, ok := <-out:
					if !ok {
						break tloop
					}

					recv++
					t.Logf("received %d", recv)
					if recv >= test.expected {
						break tloop
					}
				default:
					time.Sleep(time.Millisecond)

					initedMu.Lock()
					if test.max > 0 && inited > int(test.max) {
						t.Errorf("Scaler Exec failed: expected %d, got %d", test.max, inited)
						t.Fail()
					}
					initedMu.Unlock()

					// Release one goroutine
					release <- struct{}{}
				}
			}

			if recv != test.expected {
				t.Errorf("Scaler Exec failed: expected %d, got %d", test.expected, recv)
				t.Fail()
			}
		})
	}
}
