package result

import (
	"errors"
	"fmt"
)

func ExampleOK() {
	r := OK(123)
	if r.Ok() {
		value := r.Unwrap()
		fmt.Println(value)
	}
	// Output: 123
}

func ExampleErr() {
	r := Err[int](errors.New("error occurred"))
	if !r.Ok() {
		fmt.Println(r.Err())
	}
	// Output: error occurred
}

func ExampleResult_UnwrapOr() {
	r := Err[int](errors.New("error occurred"))
	fmt.Println(r.UnwrapOr(42))
	// Output: 42
}

// ExampleResult_UnwrapErr demonstrates the usage of the UnwrapErr function.
func ExampleResult_UnwrapErr() {
	r := Err[int](errors.New("fail"))
	if r.Ok() {
		fmt.Println(r.Unwrap())
	} else {
		fmt.Println(r.UnwrapErr())
	}
	// Output: fail
}

// ExampleResult_Expect demonstrates the usage of the Expect function.
func ExampleResult_Expect() {
	r := OK[int](123)
	fmt.Println(r.Expect("should not happen"))
	// Output: 123
}

// ExampleAndThen demonstrates the usage of AndThen function.
func ExampleAndThen() {
	r := OK("start")
	// The AndThen function can be used to chain operations on a Result.
	chainedResult := AndThen(r, func(value string) Result[string] {
		return OK(value + " chained")
	})

	// Now let's try to unwrap the chained result.
	finalValue := chainedResult.Expect("chain failed")
	fmt.Println(finalValue)
	// Output: start chained
}
