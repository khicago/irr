package irr_test

import (
	"fmt"

	"github.com/khicago/irr"
)

func ExampleError() {
	err := irr.Error("this is a new error")
	errWithParam := irr.Error("this is a new error with integer %d", 1)

	fmt.Println(err)
	fmt.Println(errWithParam)
	// Output:
	// this is a new error
	// this is a new error with integer 1
}

func ExampleWrap() {
	err := fmt.Errorf("default err message")
	wrappedErr := irr.Wrap(err, "some wrap information")
	wrappedErrWithParam := irr.Wrap(err, "some wrap information with integer %d", 1)

	fmt.Println(wrappedErr)
	fmt.Println(wrappedErrWithParam)
	// Output:
	// some wrap information, default err message
	// some wrap information with integer 1, default err message
}

func ExampleWrap_customPrint() {
	err := fmt.Errorf("default err message")
	wrappedErr := irr.Wrap(err, "some wrap information")
	wrappedErrWithParam := irr.Wrap(err, "some wrap information with integer %d", 1)

	fmt.Println(wrappedErr.ToString(false, " ==> "))
	fmt.Println(wrappedErrWithParam.ToString(false, " ==> "))
	// Output:
	// some wrap information ==> default err message
	// some wrap information with integer 1 ==> default err message
}

func ExampleTrace() {
	err := irr.Trace("this is a new error")
	errWithParam := irr.Trace("this is a new error with integer %d", 1)

	fmt.Println(err.ToString(true, ""))
	fmt.Println(errWithParam.ToString(true, ""))

	wrappedErr := irr.Track(err, "some wrap information")
	wrappedErrWithParam := irr.Track(err, "some wrap information with integer %d", 1)

	fmt.Println(wrappedErr.ToString(true, " && "))
	fmt.Println(wrappedErrWithParam.ToString(true, "\n"))
}
