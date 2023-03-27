package stream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MinWait is the absolute minimum wait time for the ticker. This is used to
// prevent the ticker from firing too often and causing too small of a wait
// time.
const MinWait = time.Millisecond

// MinLife is the minimum life time for the scaler. This is used to prevent
// the scaler from exiting too quickly, and causing too small of a lifetime.
const MinLife = time.Millisecond

// Scaler implements generic auto-scaling logic which starts with a net-zero
// set of processing routines (with the exception of the channel listener) and
// then scales up and down based on the CPU contention of a system and the speed
// at which the InterceptionFunc is able to process data. Once the incoming
// channel becomes blocked (due to nothing being sent) each of the spawned
// routines will finish out their execution of Fn and then the internal timer
// will collapse brining the routine count back to zero until there is more to
// be done.
//
// To use Scalar, simply create a new Scaler[T, U], configuring the Wait, Life,
// and InterceptFunc fields. These fields are what configure the functionality
// of the Scaler.
//
// NOTE: Fn is REQUIRED!
// Defaults: Wait = 1ns, Life = 1µs
//
// After creating the Scaler instance and configuring it, call the Exec method
// passing the appropriate context and input channel.
//
// Internally the Scaler implementation will wait for data on the incoming
// channel and attempt to send it to a layer2 channel. If the layer2 channel
// is blocking and the Wait time has been reached, then the Scaler will spawn
// a new layer2 which will increase throughput for the Scaler, and Scaler
// will attempt to send the data to the layer2 channel once more. This process
// will repeat until a successful send occurs. (This should only loop twice).
type Scaler[T, U any] struct {
	Wait time.Duration
	Life time.Duration
	Fn   InterceptFunc[T, U]

	// WaitModifier is used to modify the Wait time based on the number of
	// times the Scaler has scaled up. This is useful for systems
	// that are CPU bound and need to scale up more/less quickly.
	WaitModifier DurationScaler

	// Max is the maximum number of layer2 routines that will be spawned.
	// If Max is set to 0, then there is no limit.
	Max uint

	wScale *DurationScaler
}

var ErrFnRequired = fmt.Errorf("nil InterceptFunc, Fn is required")

// Exec starts the internal Scaler routine (the first layer of processing) and
// returns the output channel where the resulting data from the Fn function
// will be sent.
//
//nolint:funlen,gocognit // This really can't be broken up any further
func (s Scaler[T, U]) Exec(ctx context.Context, in <-chan T) (<-chan U, error) {
	ctx = _ctx(ctx)

	// set the configured tick as a pointer for execution
	s.wScale = &s.WaitModifier
	// set the original wait time on the ticker
	s.wScale.originalDuration = s.Wait

	// Fn is REQUIRED!
	if s.Fn == nil {
		return nil, ErrFnRequired
	}

	// Create outbound channel
	out := make(chan U)

	// nano-second precision really isn't feasible here, so this is arbitrary
	// because the caller did not specify a wait time. This means Scaler will
	// likely always scale up rather than waiting for an existing layer2 routine
	// to pick up data.
	if s.Wait <= MinWait {
		s.Wait = MinWait
	}

	// Minimum life of a spawned layer2 should be 1ms
	if s.Life < MinLife {
		s.Life = MinLife
	}

	go func() {
		defer close(out)

		wg := sync.WaitGroup{}
		wgMu := sync.Mutex{}

		// Ensure that the method does not close
		// until all layer2 routines have exited
		defer func() {
			wgMu.Lock()
			wg.Wait()
			wgMu.Unlock()
		}()

		l2 := make(chan T)
		ticker := time.NewTicker(s.Wait)
		defer ticker.Stop()
		step := 0
		stepMu := sync.RWMutex{}

		var max chan struct{}

		if s.Max > 0 {
			max = make(chan struct{}, s.Max)
			for i := uint(0); i < s.Max; i++ {
				max <- struct{}{}
			}
		}

	scaleLoop:
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					break scaleLoop
				}

			l2loop:
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if max != nil {
							select {
							case <-ctx.Done():
								return
							case <-max: // start a new layer2 routine
							default:
								// wait for a layer2 routine to finish
								continue l2loop
							}
						}

						wgMu.Lock()
						wg.Add(1)
						wgMu.Unlock()

						if !s.WaitModifier.inactive() {
							stepMu.Lock()
							step++
							stepMu.Unlock()
						}

						go func() {
							defer wg.Done()

							if s.Max > 0 {
								defer func() {
									select {
									case <-ctx.Done():
									case max <- struct{}{}:
									}
								}()
							}

							if !s.WaitModifier.inactive() {
								defer func() {
									stepMu.Lock()
									step--
									stepMu.Unlock()
								}()
							}

							Pipe(ctx, s.layer2(ctx, l2), out)
						}()
					case l2 <- v:
						break l2loop
					}
				}

				stepN := 0
				if !s.WaitModifier.inactive() {
					stepMu.RLock()
					stepN = step
					stepMu.RUnlock()
				}

				// Reset the ticker so that it does not immediately trip the
				// case statement on loop.
				ticker.Reset(s.wScale.scaledDuration(s.Wait, stepN))
			}
		}
	}()

	return out, nil
}

// layer2 manages the execution of the InterceptFunc. layer2 has a life time
// of s.Life and will exit if the context is canceled, the timer has reached
// its life time, or the incoming channel has been closed.
//
// If the case statement which reads from the in channel is executed, then
// layer2 will execute the Scaler function and send the result to the out
// channel. Afterward, layer2 will reset the internal timer, expanding the
// life time of the layer2, and continue to attempt another read from the in
// channel until the in channel is closed, the context is canceled, or the
// timer has reached its life time.
func (s Scaler[T, U]) layer2(ctx context.Context, in <-chan T) <-chan U {
	out := make(chan U)

	go func() {
		defer close(out)

		timer := time.NewTimer(s.Life)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				return
			case t, ok := <-in:
				if !ok {
					return
				}

				// If the function returns false, then don't send the data
				// but break out of the select statement to ensure the timer
				// is reset.
				u, send := s.Fn(ctx, t)
				if !send {
					break
				}

				// Send the resulting value to the output channel
				select {
				case <-ctx.Done():
					return
				case out <- u:
				}
			}

			// NOTE: This code is based off the doc comment for time.Timer.Stop
			// which ensures that the channel of the timer is drained before
			// resetting the timer so that it doesn't immediately trip the
			// case statement.
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(s.Life)
		}
	}()

	return out
}

// DurationScaler is used to modify the time.Duration of a ticker or timer based on
// a configured step value and modifier (between -1 and 1) value.
type DurationScaler struct {
	// Interval is the number the current step must be divisible by in order
	// to modify the time.Duration.
	Interval int

	// ScalingFactor is a value between -1 and 1 that is used to modify the
	// time.Duration of a ticker or timer. The value is multiplied by
	// the ScalingFactor is multiplied by the duration for scaling.
	//
	// For example, if the ScalingFactor is 0.5, then the duration will be
	// multiplied by 0.5. If the ScalingFactor is -0.5, then the duration will
	// be divided by 0.5. If the ScalingFactor is 0, then the duration will
	// not be modified.
	//
	// A negative ScalingFactor will cause the duration to decrease as the
	// step value increases causing the ticker or timer to fire more often
	// and create more routines. A positive ScalingFactor will cause the
	// duration to increase as the step value increases causing the ticker
	// or timer to fire less often and create less routines.
	ScalingFactor float64

	// originalDuration is the time.Duration that was passed to the
	// Scaler. This is used to reset the time.Duration of the ticker
	// or timer.
	originalDuration time.Duration

	// lastInterval is the lastInterval step that was used to modify
	// the time.Duration.
	lastInterval int
}

func (t *DurationScaler) inactive() bool {
	return t.Interval == 0 ||
		(t.ScalingFactor == 0 ||
			t.ScalingFactor <= -1 ||
			t.ScalingFactor >= 1)
}

// scaledDuration returns the modified time.Duration based on the current step (cStep).
func (t *DurationScaler) scaledDuration(
	dur time.Duration,
	currentInterval int,
) time.Duration {
	if dur < MinWait {
		dur = MinWait
	}

	if t.inactive() {
		return dur
	}

	mod := t.ScalingFactor
	if currentInterval <= t.lastInterval {
		mod = -mod
	}

	if currentInterval%t.Interval == 0 {
		t.lastInterval = currentInterval
		out := dur + time.Duration(float64(t.originalDuration)*mod)
		if out < MinWait {
			return MinWait
		}

		return out
	}

	return dur
}
