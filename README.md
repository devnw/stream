# Stream is a generic implementation for concurrency communication patterns

[![Build & Test Action Status](https://github.com/devnw/stream/actions/workflows/build.yml/badge.svg)](https://github.com/devnw/stream/actions)
[![Go Report Card](https://goreportcard.com/badge/go.atomizer.io/stream)](https://goreportcard.com/report/go.atomizer.io/stream)
[![codecov](https://codecov.io/gh/devnw/stream/branch/main/graph/badge.svg)](https://codecov.io/gh/devnw/stream)
[![Go Reference](https://pkg.go.dev/badge/go.atomizer.io/stream.svg)](#documentation)
[![License: Apache 2.0](https://img.shields.io/badge/license-Apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

Stream provides a set of generic functions for working concurrent
design patterns in Go.

## Installation

To install the package, run:

```bash
    go get -u go.atomizer.io/stream@latest
```

## Importing

It is recommended to use the package via the following import:

`import . "go.atomizer.io/stream"`

Using the `.` import allows for functions to be called directly as if the
functions were in the same namespace without the need to append the package
name.

## Documentation

### TYPES

**`type InterceptFunc[T, U any] func(context.Context, T) (U, bool)`**

```go
type Scaler[T, U any] struct {
        Wait time.Duration
        Life time.Duration
        Fn   InterceptFunc[T, U]
}
```

>Scaler implements generic auto-scaling logic which starts with a net-zero
    set of processing routines (with the exception of the channel listener) and
    then scales up and down based on the CPU contention of a system and the
    speed at which the InterceptionFunc is able to process data. Once the
    incoming channel becomes blocked (due to nothing being sent) each of the
    spawned routines will finish out their execution of Fn and then the internal
    timer will collapse brining the routine count back to zero until there is
    more to be done.
>To use Scalar, simply create a new Scaler[T, U], configuring the Wait, Life,
    and InterceptFunc fields. These fields are what configure the functionality
    of the Scaler.
>NOTE: Fn is REQUIRED!
>After creating the Scaler instance and configuring it, call the Exec method
    passing the appropriate context and input channel.
>Internally the Scaler implementation will wait for data on the incoming
    channel and attempt to send it to a layer2 channel. If the layer2 channel is
    blocking and the Wait time has been reached, then the Scaler will spawn a
    new layer2 which will increase throughput for the Scaler, and Scaler will
    attempt to send the data to the layer2 channel once more. This process will
    repeat until a successful send occurs. (This should only loop twice)

**`func (s Scaler[T, U]) Exec(ctx context.Context, in <-chan T) (<-chan U, error)`**
    Exec starts the internal Scaler routine (the first layer of processing) and
    returns the output channel where the resulting data from the Fn function
    will be sent.

### FUNCTIONS

**`func Distribute[T any](ctx context.Context, in <-chan T, out ...chan<- T)`**
>Distribute accepts an incoming data channel and distributes the data among
    the supplied outgoing data channels. This distribution is done
    stochastically using the cryptographic random number generator.
>**NOTE:** Execute the Distribute function in a goroutine if parallel execution
    is desired. Cancelling the context or closing the incoming channel is
    important to ensure that the goroutine is properly terminated.

**`func FanIn[T any](ctx context.Context, in ...<-chan T) <-chan T`**
>FanIn accepts incoming data channels and forwards returns a single channel
    that receives all the data from the supplied channels.
>**NOTE:** The transfer takes place in a goroutine for each channel so ensuring
    that the context is cancelled or the incoming channels are closed is
    important to ensure that the goroutine is terminated.

**`func FanOut[T any](ctx context.Context, in <-chan T, out ...chan<- T)`**
>FanOut accepts an incoming data channel and forwards the data from it to the
    supplied outgoing data channels.
>**NOTE:** Execute the FanOut function in a goroutine if parallel execution is
    desired. Cancelling the context or closing the incoming channel is important
    to ensure that the goroutine is properly terminated.

**`func Intercept[T, U any](ctx context.Context,in <-chan T,fn InterceptFunc[T,
U]) <-chan U`**
>Intercept accepts an incoming data channel and a function literal that
    accepts the incoming data and returns data of the same type and a boolean
    indicating whether the data should be forwarded to the output channel. The
    function is executed for each data item in the incoming channel as long as
    the context is not cancelled or the incoming channel remains open.

**`func Pipe[T any](ctx context.Context, in <-chan T, out chan<- T)`**
>Pipe accepts an incoming data channel and pipes it to the supplied outgoing
    data channel.
>**NOTE:** Execute the Pipe function in a goroutine if parallel execution is
    desired. Cancelling the context or closing the incoming channel is important
    to ensure that the goroutine is properly terminated.

**`func ToStream[U ~[]T, T any](ctx context.Context, in U) <-chan T`**
>ToStream accepts an slice of values and converts them to a channel.
>**NOTE:** This function does NOT use a buffered channel.

## Benchmarks

To execute the benchmarks, run the following command:

```bash
    go test -bench=. ./...
```

To view benchmarks over time for the `main` branch of the repository they can
be seen on our [Benchmark Report Card].

[Benchmark Report Card]: https://go.devnw.com/stream/dev/bench/
