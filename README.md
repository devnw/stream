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

### FUNCTIONS

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
