# irr

Irr is an error library based on the handling stack.
It supports error wrapping, handling-stack tracing, and error stack traversal.

## Table of Contents

- [Overview](#overview)
- [Concept](#concept)
- [Usage](#usage)
  - [Import package](#import-package)
  - [Basic Usage](#basic-usage)
  - [Work with errors' stack trace](#work-with-errors-stack-trace)
- [TODO: Not finished yet](#todo-not-finished-yet)

## Overview

irr provides

- Error wrapping
- Error link traversal
- Optional stack tracing

## Concept

Irr is more concerned with the error handling-stack than the usual call stack.

The handling-stack is somewhat similar to the function call stack, but it is more reflective of the relationship between the flow before and after the exception handling than the function call relationship. Therefore, it is more reflective of the actual handling of the logical function library.

For example, when function A calls function B on line $l_a$, and an error is generated on line $l_b$ of function B.
The usual error tracing, gives a tuple of $<A,l_a>$ $<B, l_b>$

In fact, the exception errB returned by B is often not handled in $l_a$, but is distributed to some subsequent logic, or even to other sub-functions. Therefore, the function call stack can only focus on the generation of exceptions, but in many cases, we need to focus on the whole logic chain from the generation of exceptions to their final processing.

Handling-stack is usually done via wrap. In irr handling, the advantages of stack trace and error wrap are combined, so that there is trace information for both exceptions and handling sessions. This makes it easy to trace the call chain.

To avoid redundant information and unnecessary performance overhead, irr does not advocate exporting the call stack in non-error situations, and therefore only provides methods related to error handling. In addition, irr advocates that developers should handle exceptions clearly, so it provides methods to skip the stack frame, or just warp without outputting the stack, to serve developers with better error handling practices.

## Usage

### Import package

```go
import (
    "github.com/khicago/irr"
)
```

### Basic Usage

Create an error

```go
err := irr.Error("this is a new error")
errWithParam :=  irr.Error("this is a new error with integer %d", 1)
```

if you print them, you will got

```go
fmt.Println(err)
fmt.Println(errWithParam)
// Output:
// this is a new error
// this is a new error with integer 1
```

Or you can easilly wrap an error

```go
wrappedErr := irr.Wrap(err, "some wrap information")
wrappedErrWithParam := irr.Wrap(err, "some wrap information with integer %d", 1)

fmt.Println(wrappedErr)
fmt.Println(wrappedErrWithParam)
// when err := fmt.Errorf("default err message"), the outputs will be
// Output:
// some wrap information; default err message
// some wrap information with integer 1; default err message
```

and you can define the output format by yourself

```go
fmt.Println(wrappedErr.ToString(false, " ==> "))
fmt.Println(wrappedErrWithParam.ToString(false, " ==> "))
// Output:
// some wrap information ==> default err message
// some wrap information with integer 1 ==> default err message
```

### Work with errors' stack trace

Create an error with stack trace

```go
err := irr.Trace("this is a new error")
errWithParam :=  irr.Trace("this is a new error with integer %d", 1)
```

By default, the trace info will not be print by `Error()` method
ToString method can be used to print trace info

```go
fmt.Println(err.ToString(true, ""))
fmt.Println(errWithParam.ToString(true, ""))
// this is a new error your_function@/.../your_code.go:line
// this is a new error with integer 1 your_function@/.../your_code.go:line
```

You can also easilly wrap an error with stack trace

```go
wrappedErr := irr.Track(err, "some wrap information")
wrappedErrWithParam := irr.Track(err, "some wrap information with integer %d", 1)
```

The result can be exported in the same way, and you can set the splitor of each stack.

```go
fmt.Println(wrappedErr.ToString(true, " && "))
// some wrap information your_outer_function@/.../your_outer_code.go:line && this is a new error your_function@/.../your_code.go:line
fmt.Println(wrappedErrWithParam.ToString(true, "\n"))
// some wrap information with integer 1 your_outer_function@/.../your_outer_code.go:line
// this is a new error your_function@/.../your_code.go:line
```

## TODO: Not finished yet
