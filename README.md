# Stream is a generic implementation for concurrency communication patterns

[![Build & Test Action Status](https://github.com/devnw/stream/actions/workflows/build.yml/badge.svg)](https://github.com/devnw/stream/actions)
[![Go Report Card](https://goreportcard.com/badge/go.atomizer.io/stream)](https://goreportcard.com/report/go.atomizer.io/stream)
[![codecov](https://codecov.io/gh/devnw/stream/branch/main/graph/badge.svg)](https://codecov.io/gh/devnw/stream)
[![Go Reference](https://pkg.go.dev/badge/go.atomizer.io/stream.svg)](https://pkg.go.dev/go.atomizer.io/stream)
[![License: Apache 2.0](https://img.shields.io/badge/license-Apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

Stream provides a set of generic functions for working concurrent
design patterns in Go.

## Installation

To install the package, run:

```bash
    go get -u go.atomizer.io/stream@latest
```

## Usage

```go
    import "go.atomizer.io/stream"
```

## Benchmarks

To execute the benchmarks, run the following command:

```bash
    go test -bench=. ./...
```

To view benchmarks over time for the `main` branch of the repository they can
be seen on our [Benchmark Report Card].

[Benchmark Report Card]: https://devnw.github.io/stream/dev/bench/

<!-- GODOC_START -->
