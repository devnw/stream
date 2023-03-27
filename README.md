# stream
--
    import "."

Package stream provides a set of generic functions for working concurrent design
patterns in Go.

[![Build & Test Action
Status](https://github.com/devnw/stream/actions/workflows/build.yml/badge.svg)](https://github.com/devnw/stream/actions)
[![Go Report
Card](https://goreportcard.com/badge/go.atomizer.io/stream)](https://goreportcard.com/report/go.atomizer.io/stream)
[![codecov](https://codecov.io/gh/devnw/stream/branch/main/graph/badge.svg)](https://codecov.io/gh/devnw/stream)
[![Go
Reference](https://pkg.go.dev/badge/go.atomizer.io/stream.svg)](https://pkg.go.dev/go.atomizer.io/stream)
[![License: Apache
2.0](https://img.shields.io/badge/license-Apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![PRs
Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

## Installation

To install the package, run:

    go get -u go.atomizer.io/stream@latest

## Usage

    import "go.atomizer.io/stream"

## Benchmarks

To execute the benchmarks, run the following command:

    go test -bench=. ./...

To view benchmarks over time for the `main` branch of the repository they can be
seen on our [Benchmark Report Card].

[Benchmark Report Card]: https://devnw.github.io/stream/dev/bench/

## Usage

```go
var ErrFnRequired = fmt.Errorf("nil InterceptFunc, Fn is required")
```

#### func  Any

```go
func Any[T any](ctx context.Context, in <-chan T) <-chan any
```
Any accepts an incoming data channel and converts the channel to a readonly
channel of the `any` type.

#### func  Distribute

```go
func Distribute[T any](
	ctx context.Context, in <-chan T, out ...chan<- T,
)
```
Distribute accepts an incoming data channel and distributes the data among the
supplied outgoing data channels using a dynamic select statement.

NOTE: Execute the Distribute function in a goroutine if parallel execution is
desired. Canceling the context or closing the incoming channel is important to
ensure that the goroutine is properly terminated.

#### func  Drain

```go
func Drain[T any](ctx context.Context, in <-chan T)
```
Drain accepts a channel and drains the channel until the channel is closed or
the context is canceled.

#### func  FanIn

```go
func FanIn[T any](ctx context.Context, in ...<-chan T) <-chan T
```
FanIn accepts incoming data channels and forwards returns a single channel that
receives all the data from the supplied channels.

NOTE: The transfer takes place in a goroutine for each channel so ensuring that
the context is canceled or the incoming channels are closed is important to
ensure that the goroutine is terminated.

#### func  FanOut

```go
func FanOut[T any](
	ctx context.Context, in <-chan T, out ...chan<- T,
)
```
FanOut accepts an incoming data channel and copies the data to each of the
supplied outgoing data channels.

NOTE: Execute the FanOut function in a goroutine if parallel execution is
desired. Canceling the context or closing the incoming channel is important to
ensure that the goroutine is properly terminated.

#### func  Intercept

```go
func Intercept[T, U any](
	ctx context.Context,
	in <-chan T,
	fn InterceptFunc[T, U],
) <-chan U
```
Intercept accepts an incoming data channel and a function literal that accepts
the incoming data and returns data of the same type and a boolean indicating
whether the data should be forwarded to the output channel. The function is
executed for each data item in the incoming channel as long as the context is
not canceled or the incoming channel remains open.

#### func  Pipe

```go
func Pipe[T any](
	ctx context.Context, in <-chan T, out chan<- T,
)
```
Pipe accepts an incoming data channel and pipes it to the supplied outgoing data
channel.

NOTE: Execute the Pipe function in a goroutine if parallel execution is desired.
Canceling the context or closing the incoming channel is important to ensure
that the goroutine is properly terminated.

#### type DurationScaler

```go
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
}
```

DurationScaler is used to modify the time.Duration of a ticker or timer based on
a configured step value and modifier (between -1 and 1) value.

#### type InterceptFunc

```go
type InterceptFunc[T, U any] func(context.Context, T) (U, bool)
```


#### type Scaler

```go
type Scaler[T, U any] struct {
	Wait time.Duration
	Life time.Duration
	Fn   InterceptFunc[T, U]

	// WaitModifier is used to modify the Wait time based on the number of
	// times the Scaler has scaled up. This is useful for systems
	// that are CPU bound and need to scale up more/less quickly.
	WaitModifier DurationScaler
}
```

Scaler implements generic auto-scaling logic which starts with a net-zero set of
processing routines (with the exception of the channel listener) and then scales
up and down based on the CPU contention of a system and the speed at which the
InterceptionFunc is able to process data. Once the incoming channel becomes
blocked (due to nothing being sent) each of the spawned routines will finish out
their execution of Fn and then the internal timer will collapse brining the
routine count back to zero until there is more to be done.

To use Scalar, simply create a new Scaler[T, U], configuring the Wait, Life, and
InterceptFunc fields. These fields are what configure the functionality of the
Scaler.

NOTE: Fn is REQUIRED!

After creating the Scaler instance and configuring it, call the Exec method
passing the appropriate context and input channel.

Internally the Scaler implementation will wait for data on the incoming channel
and attempt to send it to a layer2 channel. If the layer2 channel is blocking
and the Wait time has been reached, then the Scaler will spawn a new layer2
which will increase throughput for the Scaler, and Scaler will attempt to send
the data to the layer2 channel once more. This process will repeat until a
successful send occurs. (This should only loop twice).

#### func (Scaler[T, U]) Exec

```go
func (s Scaler[T, U]) Exec(ctx context.Context, in <-chan T) (<-chan U, error)
```
Exec starts the internal Scaler routine (the first layer of processing) and
returns the output channel where the resulting data from the Fn function will be
sent.
