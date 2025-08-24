# IRR - Structured Error Handling for Go

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Build Status](https://travis-ci.org/khicago/irr.svg?branch=master)](https://travis-ci.org/khicago/irr)
[![codecov](https://codecov.io/gh/khicago/irr/branch/master/graph/badge.svg)](https://codecov.io/gh/khicago/irr)
[![Go Report Card](https://goreportcard.com/badge/github.com/khicago/irr)](https://goreportcard.com/report/github.com/khicago/irr)
[![GoDoc](https://godoc.org/github.com/khicago/irr?status.svg)](https://godoc.org/github.com/khicago/irr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

IRR is a Go error handling library that preserves context information across error propagation chains. It addresses the common problem where error messages lose critical debugging information as they bubble up through application layers.

**[中文版本 (Chinese Version)](README_zh.md)**

## Problem Statement

Traditional Go error handling often results in information loss:

```go
// Traditional approach - loses context at each layer
func processUser(id string) error {
    user, err := getUserFromDB(id)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }
    
    if err := validateUser(user); err != nil {
        return fmt.Errorf("validation failed: %w", err) 
    }
    
    return nil
}
// Result: \"validation failed: failed to get user: connection timeout\"
// Missing: when, where, what context, error classification
```

## Solution

IRR provides structured error handling with context preservation:

```go
// IRR approach - preserves structured context
func processUser(ctx context.Context, id string) error {
    user, err := getUserFromDB(id)
    if err != nil {
        return irr.Wraps(err, \"failed to get user for id=%s\", id).
            Tag(\"operation\", \"user_fetch\").
            Tag(\"user_id\", id).
            WithEnricher(&enrichers.TimeoutEnricher{}).
            BuildWithContext(ctx)
    }
    
    if err := validateUser(user); err != nil {
        return irr.Wraps(err, \"user validation failed\").
            Tag(\"operation\", \"validation\").
            Tag(\"user_id\", id).
            BuildWithContext(ctx)
    }
    
    return nil
}
// Result: Structured error with tags, context, and debugging information
```

## Key Features

### Builder Pattern API
Fluent interface for constructing errors with optional attributes:
```go
err := irr.New(\"operation failed\").
    Code(1001).
    Tag(\"service\", \"user\").
    Trace().
    Build()
```

### Context Enrichment
Pluggable enrichers extract information from Go contexts:
```go
import \"github.com/khicago/irr/enrichers\"

err := irr.New(\"timeout occurred\").
    WithEnricher(enrichers.NewRequestInfoEnricher()). // extracts trace_id, user_id
    WithEnricher(&enrichers.TimeoutEnricher{}).       // detects deadline exceeded
    BuildWithContext(ctx)
```

### Enterprise Error Codes
Structured error classification with IRC (IRR Code) system:
```go
const (
    ErrSystemDatabase    irc.Code = 1001  // System errors (1000-1999)
    ErrBusinessValidation irc.Code = 2001  // Business errors (2000-2999)
    ErrAPINotFound       irc.Code = 3002  // API errors (3000-3999)
)

return ErrBusinessValidation.Error(\"email is required\")
```

### Optional Monitoring Integration
Non-intrusive metrics collection at application boundaries:
```go
// At handler layer
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    err := h.businessLogic(r.Context())
    if err != nil {
        h.analyzer.Analyze(err)  // Optional metrics collection
        h.logger.Error(err.Details())
        http.Error(w, \"Internal server error\", 500)
    }
}
```

## Installation

```bash
go get github.com/khicago/irr
```

## Basic Usage

### Creating Errors

```go
// Simple error
err := irr.New(\"user not found\").Build()

// Error with metadata
err := irr.New(\"invalid input: %s\", input).
    Code(400).
    Tag(\"field\", \"email\").
    Tag(\"service\", \"user\").
    Build()

// Error with stack trace
err := irr.New(\"critical failure\").
    Trace().
    Build()
```

### Wrapping Errors

```go
user, err := db.GetUser(id)
if err != nil {
    return irr.Wraps(err, \"failed to fetch user\").
        Tag(\"user_id\", id).
        Tag(\"operation\", \"get_user\").
        BuildWithContext(ctx)
}
```

### Context-Aware Errors

```go
// Automatic timeout detection
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

err := irr.New(\"operation failed\").
    WithEnricher(&enrichers.TimeoutEnricher{}).
    BuildWithContext(ctx)

// If context deadline is exceeded:
// err.Tags()[\"timeout\"] = \"true\"
// err.Tags()[\"deadline_exceeded\"] = \"2024-01-01T10:05:00Z\"
```

### Enterprise Error Codes

```go
import \"github.com/khicago/irr/codes\"

// Define error taxonomy
const (
    ErrUserNotFound     codes.Code = 2001
    ErrInvalidInput     codes.Code = 2002
    ErrDatabaseTimeout  codes.Code = 1001
)

// Create classified errors
func GetUser(id string) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, ErrDatabaseTimeout.Track(err, \"query failed for user=%s\", id)
    }
    
    if user == nil {
        return nil, ErrUserNotFound.Error(\"user id=%s not found\", id)
    }
    
    return user, nil
}
```

### Result Type (Optional)

Rust-inspired error handling for functional programming style:

```go
import \"github.com/khicago/irr/result\"

func safeDivision(a, b int) result.Result[int] {
    if b == 0 {
        return result.Err[int](irr.New(\"division by zero\").Build())
    }
    return result.Ok(a / b)
}

// Usage
res := safeDivision(10, 2)
if res.IsOk() {
    fmt.Printf(\"Result: %d\", res.Unwrap())
} else {
    fmt.Printf(\"Error: %v\", res.UnwrapErr())
}
```

## Advanced Usage

### Custom Enrichers

```go
type DatabaseEnricher struct{}

func (d *DatabaseEnricher) Enrich(ctx context.Context, err irr.IRR) irr.IRR {
    if dbName := ctx.Value(\"db-name\"); dbName != nil {
        err.SetTag(\"database\", dbName.(string))
    }
    return err
}

// Usage
err := irr.New(\"query failed\").
    WithEnricher(&DatabaseEnricher{}).
    BuildWithContext(ctx)
```

### Handler-Level Error Processing

```go
type ErrorHandler struct {
    logger   *zap.Logger
    analyzer *metrics.ErrorAnalyzer
}

func (h *ErrorHandler) HandleError(err error) {
    if irrErr, ok := err.(irr.IRR); ok {
        // Extract structured information
        code := irrErr.Code()
        tags := irrErr.Tags()
        
        // Log with structure
        h.logger.Error(\"operation failed\",
            zap.Error(err),
            zap.Int64(\"code\", code),
            zap.Any(\"tags\", tags),
        )
        
        // Optional metrics (non-blocking)
        go h.analyzer.Analyze(irrErr)
    }
}
```

### Cross-Service Error Propagation

```go
// Service A
err := irr.New(\"service unavailable\").
    Tag(\"service\", \"user-service\").
    Tag(\"trace_id\", ctx.Value(\"trace_id\")).
    Code(5001).
    BuildWithContext(ctx)

// Serialize for transmission
errorJSON := err.MarshalJSON()

// Service B - reconstruct error context
reconstructedErr := irr.UnmarshalFromJSON(errorJSON)

// Chain with new context
chainedErr := irr.Wraps(reconstructedErr, \"downstream service failed\").
    Tag(\"service\", \"order-service\").
    BuildWithContext(ctx)
```

## Performance

IRR is designed for production use with minimal overhead:

```
BenchmarkError-8           2000000    750 ns/op    112 B/op    2 allocs/op
BenchmarkWrap-8            1500000    950 ns/op    144 B/op    3 allocs/op  
BenchmarkTrace-8           1000000   1200 ns/op    256 B/op    4 allocs/op
BenchmarkTrack-8            800000   1450 ns/op    288 B/op    5 allocs/op

// Comparison with standard library:
BenchmarkStdError-8        3000000    420 ns/op     64 B/op    1 allocs/op
BenchmarkStdWrap-8         2000000    680 ns/op     96 B/op    2 allocs/op
```

**Performance features:**
- Memory pooling for trace objects
- Lazy string building with caching
- Zero-allocation fast paths for simple cases
- Efficient atomic operations for metrics

## Quality Assurance

**Test Coverage:**
```
Package                              Coverage
github.com/khicago/irr               88.7%
github.com/khicago/irr/internal/errors  93.3%
github.com/khicago/irr/enrichers     100.0%
github.com/khicago/irr/reporter      86.9%
github.com/khicago/irr/codes         71.4%
github.com/khicago/irr/metrics       70.7%
```

**Quality Standards:**
- Semantic test naming for maintainability
- Edge case and concurrent access testing
- Benchmark tests for performance regression detection
- Complete API contract validation

## Comparison with Other Libraries

| Feature | IRR | pkg/errors | std errors | zerolog |
|---------|-----|------------|------------|---------|
| Structured Context | Yes | No | No | Limited |
| Error Codes | Yes (codes) | No | No | No |
| Context Integration | Native | No | No | No |
| Monitoring Ready | Yes | No | No | Partial |
| Result Type | Optional | No | No | No |
| Performance | Optimized | Good | Excellent | Good |

## Documentation

- [API Documentation](https://godoc.org/github.com/khicago/irr)
- [Architecture Design](./design.0.2.md) - In-depth design analysis and layer-by-layer breakdown
- [IRC Enterprise Guide](./docs/irc-enterprise-practices.md) - Production patterns and best practices
- [Examples](./examples/) - Practical usage examples

## Migration

### From Standard Library

```go
// Before
return fmt.Errorf(\"operation failed: %w\", err)

// After
return irr.Wraps(err, \"operation failed\").
    Tag(\"operation\", \"user_fetch\").
    BuildWithContext(ctx)
```

### From pkg/errors

```go
// Before
return errors.Wrap(err, \"operation failed\")

// After  
return irr.Wraps(err, \"operation failed\").
    Trace().
    BuildWithContext(ctx)
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Run tests: `go test -v ./...`
4. Run benchmarks: `go test -bench=. -benchmem`
5. Submit a pull request

**Development Setup:**
```bash
git clone https://github.com/khicago/irr.git
cd irr
go mod tidy
go test -v ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Rust's structured error handling patterns
- Built for Go community's need for better error context preservation
- Thanks to all contributors and early adopters

---

**Used by companies and projects worldwide for production error handling.**