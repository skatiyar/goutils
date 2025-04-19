# goutils
Golang utility functions to ease working with goroutines and provide functional programming utilities. This package is inspired by [async](https://caolan.github.io/async) and implements utility functions in an idiomatic Go way.

[![Go Report Card](https://goreportcard.com/badge/github.com/skatiyar/goutils)](https://goreportcard.com/report/github.com/skatiyar/goutils)
[![Go Reference](https://pkg.go.dev/badge/github.com/skatiyar/goutils.svg)](https://pkg.go.dev/github.com/skatiyar/goutils)
[![codecov](https://codecov.io/gh/skatiyar/goutils/graph/badge.svg?token=ND6O9OWB1H)](https://codecov.io/gh/skatiyar/goutils)

## Features
goutils provides around 60 functions that include the usual 'functional' suspects (map, reduce, filter, each…) as well as some common patterns for asynchronous control flow (async/await, queue, waterfall). Package [`goutils`](https://pkg.go.dev/github.com/skatiyar/goutils) has sync `map`, `reduce` etc., whereas [`goutils/async`](https://pkg.go.dev/github.com/skatiyar/goutils/async) has concurrent `map`, `reduce` etc.

### Example: Using [`async.Async`](https://pkg.go.dev/github.com/skatiyar/goutils/async#Async)

The `async.Async` function allows you to execute multiple functions concurrently and wait for all of them to complete.

```go
package main

import (
    "fmt"
    "time"

    "github.com/skatiyar/goutils/async"
)

func taskOne () (string, error) {
    time.Sleep(time.Second)
    return "Task 1", nil
}

func taskTwo () (string, error) {
    time.Sleep(2 * time.Second)
    panic("Task 2 paniced!")
}

func taskThree () (string, error) {
    time.Sleep(3 * time.Second)
    return "Task 3", nil
}

func main() {
    t1Result := async.Async(taskOne)
    t2Result := async.Async(taskTwo)
    t3Result := async.Async(taskThree)

    t2Data, t2Error := t2Result.Await()
    if t2Error != nil {
        fmt.Printf("Error: %v\n", t2Error)
    } else {
        fmt.Println("Task completed successfully", t2Data)
    }

    t1Data, t1Error := t1Result.Await()
    if t1Error != nil {
        fmt.Printf("Error: %v\n", t1Error)
    } else {
        fmt.Println("Task completed successfully", t1Data)
    }

    t3Data, t3Error := t3Result.Await()
    if t3Error != nil {
        fmt.Printf("Error: %v\n", t3Error)
    } else {
        fmt.Println("Task completed successfully", t3Data)
    }
}
```

Output:
```sh
Error: panic in go routine
Task completed successfully Task 1
Task completed successfully Task 3
```


## Installation

To install the package, run:

```bash
go get github.com/skatiyar/goutils
```

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
