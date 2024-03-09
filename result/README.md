# Result Package for Go

The `result` package provides a robust way to handle errors in Go by emulating the `Result` type found in languages like Rust. This enables Go developers to write cleaner, more expressive error handling code with clear distinction between successful and error states.

## Features

- Type-safe results with `OK` and `Err` constructors
- Chainable error handling using `AndThen`
- Methods like `Unwrap`, `UnwrapOr`, and `Expect` for convenient value extraction
- Panic-based error unwrapping to clearly delineate error cases

## Installation

To use the `result` package in your Go project, install it using `go get`:

```sh
go get -u github.com/khicago/irr/result
```
Replace your_username with your GitHub username or organization where the package is hosted.

## Usage

### Creating a Result

```go
import "github.com/khicago/irr/result"

// Create a Result with a successful value
okResult := result.OK[int](42)

// Create a Result with an error
errResult := result.Err[int](errors.New("some error"))
```

### Handling a Result

Use `Ok` to check for success and handle values appropriately:

```go
if val.Ok() {
    value := ok.Unwrap()
    fmt.Println(value)
} else {
    e := ok.UnwrapErr()
    fmt.Println(e)
}
```

or using `Match`

```go
switch val, err := val.Match() {
case err != nil: ...
default: ...
}
```

### Chaining Results

Chain operations that may fail with AndThen:

```go
v1 := result.AndThen(ok, func(value int) result.Result[string] {
    return result.OK(fmt.Sprintf("Processed: %d", value))
})
v2 := result.AndThen(v1, func(value string) result.Result[string] {
    return result.OK(fmt.Sprintf("+: %s", value))
})
...

if !vn.Ok() {
	...
}

fmt.Println(chainedResult.Unwrap())
```

### Forcing Value Extraction

Extract values with UnwrapOr and Expect, noting that these methods can panic:

```go
// Returns the contained value or a default
defaultValue := err.UnwrapOr(42)

// Panics if the result is an error
safeValue := ok.Expect("should not panic")
```