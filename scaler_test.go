package stream

import (
	"context"
	"testing"
	"time"

	"go.structs.dev/gen"
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
			//nolint:staticcheck
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

	//nolint:staticcheck
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
	//nolint:staticcheck
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
	//nolint:staticcheck
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
