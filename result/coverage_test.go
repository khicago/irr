package result

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResult_ComprehensiveCoverage 测试所有未覆盖的方法和边缘情况
func TestResult_ComprehensiveCoverage(t *testing.T) {
	t.Run("IsOk and IsErr methods", func(t *testing.T) {
		okResult := OK(42)
		errResult := Err[int, error](errors.New("test error"))

		// Test IsOk
		assert.True(t, okResult.IsOk())
		assert.False(t, errResult.IsOk())

		// Test IsErr
		assert.False(t, okResult.IsErr())
		assert.True(t, errResult.IsErr())
	})

	t.Run("Ok and Err methods", func(t *testing.T) {
		okResult := OK(42)
		errResult := Err[int, error](errors.New("test error"))

		// Test Ok method (alias for IsOk)
		assert.True(t, okResult.Ok())
		assert.False(t, errResult.Ok())

		// Test Err method
		var zeroErr error
		assert.Equal(t, zeroErr, okResult.Err()) // Should return zero value
		assert.NotNil(t, errResult.Err())
	})

	t.Run("Value and Error methods", func(t *testing.T) {
		okResult := OK(42)
		errResult := Err[int, error](errors.New("test error"))

		// Test Value method
		value, hasValue := okResult.Value()
		assert.True(t, hasValue)
		assert.Equal(t, 42, value)

		value2, hasValue2 := errResult.Value()
		assert.False(t, hasValue2)
		assert.Equal(t, 0, value2) // Zero value for int

		// Test Error method
		err, hasError := okResult.Error()
		assert.False(t, hasError)
		assert.Nil(t, err)

		err2, hasError2 := errResult.Error()
		assert.True(t, hasError2)
		assert.Contains(t, err2.Error(), "test error")
	})

	t.Run("Unwrap edge cases", func(t *testing.T) {
		// Test Unwrap with zero value
		okResult := OK(0)
		assert.Equal(t, 0, okResult.Unwrap())

		// Test Unwrap with empty string
		okStringResult := OK("")
		assert.Equal(t, "", okStringResult.Unwrap())

		// Test Unwrap with nil pointer
		var nilPtr *int
		okPtrResult := OK(nilPtr)
		assert.Nil(t, okPtrResult.Unwrap())
	})

	t.Run("UnwrapOr edge cases", func(t *testing.T) {
		// Test UnwrapOr with Ok result and default
		okResult := OK(42)
		assert.Equal(t, 42, okResult.UnwrapOr(999))

		// Test UnwrapOr with Err result and zero default
		errResult := Err[int, error](errors.New("error"))
		assert.Equal(t, 0, errResult.UnwrapOr(0))

		// Test UnwrapOr with different types
		okStringResult := OK("hello")
		assert.Equal(t, "hello", okStringResult.UnwrapOr("default"))

		errStringResult := Err[string, error](errors.New("error"))
		assert.Equal(t, "default", errStringResult.UnwrapOr("default"))
	})

	t.Run("UnwrapErr edge cases", func(t *testing.T) {
		// Test UnwrapErr with nil error
		var nilErr error
		errResult := Err[int, error](nilErr)
		assert.Nil(t, errResult.UnwrapErr())

		// Test UnwrapErr with specific error type
		customErr := errors.New("custom error")
		errResult2 := Err[int, error](customErr)
		assert.Equal(t, customErr, errResult2.UnwrapErr())

		// Test UnwrapErr on Ok result (should panic)
		okResult := OK(42)
		assert.Panics(t, func() {
			okResult.UnwrapErr()
		})
	})

	t.Run("Expect method", func(t *testing.T) {
		// Test Expect with Ok result
		okResult := OK(42)
		assert.Equal(t, 42, okResult.Expect("should not panic"))

		// Test Expect with Err result (should panic)
		errResult := Err[int, error](errors.New("test error"))
		assert.Panics(t, func() {
			errResult.Expect("this should panic")
		})
	})

	t.Run("Match with complex scenarios", func(t *testing.T) {
		// Test Match with Ok result
		okResult := OK(5)
		value, err := okResult.Match()
		assert.Equal(t, 5, value)
		assert.Nil(t, err)

		// Test Match with Err result
		testErr := errors.New("test error")
		errResult := Err[int, error](testErr)
		value2, err2 := errResult.Match()
		assert.Equal(t, 0, value2) // Zero value for int
		assert.Equal(t, testErr, err2)

		// Test Match with string type
		okStringResult := OK("hello")
		strValue, strErr := okStringResult.Match()
		assert.Equal(t, "hello", strValue)
		assert.Nil(t, strErr)

		errStringResult := Err[string, error](testErr)
		strValue2, strErr2 := errStringResult.Match()
		assert.Equal(t, "", strValue2) // Zero value for string
		assert.Equal(t, testErr, strErr2)
	})

	t.Run("MatchFunc method", func(t *testing.T) {
		// Test MatchFunc with Ok result
		var called string
		okResult := OK(10)
		okResult.MatchFunc(
			func(value int) {
				called = "Got value: " + string(rune(value+'0'))
			},
			func(err error) {
				called = "Got error: " + err.Error()
			},
		)
		assert.Contains(t, called, "Got value")

		// Test MatchFunc with Err result
		var called2 string
		errResult := Err[int, error](errors.New("test error"))
		errResult.MatchFunc(
			func(value int) {
				called2 = "Should not reach here"
			},
			func(err error) {
				called2 = "Error: " + err.Error()
			},
		)
		assert.Contains(t, called2, "Error: test error")
	})

	t.Run("Complex type handling", func(t *testing.T) {
		// Test with struct types
		type Person struct {
			Name string
			Age  int
		}

		person := Person{Name: "Alice", Age: 30}
		okResult := OK(person)
		assert.True(t, okResult.IsOk())
		unwrapped := okResult.Unwrap()
		assert.Equal(t, "Alice", unwrapped.Name)
		assert.Equal(t, 30, unwrapped.Age)

		// Test with pointer types
		ptrResult := OK(&person)
		assert.True(t, ptrResult.IsOk())
		unwrappedPtr := ptrResult.Unwrap()
		assert.Equal(t, "Alice", unwrappedPtr.Name)

		// Test with slice types
		slice := []int{1, 2, 3}
		sliceResult := OK(slice)
		assert.True(t, sliceResult.IsOk())
		unwrappedSlice := sliceResult.Unwrap()
		assert.Equal(t, []int{1, 2, 3}, unwrappedSlice)
	})

	t.Run("Constructors comprehensive", func(t *testing.T) {
		// Test different constructor functions

		// Ok generic constructor
		genericOk := Ok[int, error](42)
		assert.True(t, genericOk.IsOk())
		assert.Equal(t, 42, genericOk.Unwrap())

		// Err generic constructor
		genericErr := Err[int, error](errors.New("error"))
		assert.True(t, genericErr.IsErr())

		// OkError constructor
		okError := OkError(42)
		assert.True(t, okError.IsOk())

		// ErrError constructor
		errError := ErrError[int](errors.New("error"))
		assert.True(t, errError.IsErr())

		// OK constructor (alias)
		okAlias := OK(42)
		assert.True(t, okAlias.IsOk())
	})

	t.Run("ToTuple and FromTuple", func(t *testing.T) {
		// Test ToTuple with Ok result
		okResult := OK(42)
		value, err := ToTuple(okResult)
		assert.Equal(t, 42, value)
		assert.Nil(t, err)

		// Test ToTuple with Err result
		errResult := Err[int, error](errors.New("test error"))
		value2, err2 := ToTuple(errResult)
		assert.Equal(t, 0, value2)
		assert.Contains(t, err2.Error(), "test error")

		// Test FromTuple with no error
		result3 := FromTuple(99, nil)
		assert.True(t, result3.IsOk())
		assert.Equal(t, 99, result3.Unwrap())

		// Test FromTuple with error
		result4 := FromTuple(0, errors.New("conversion error"))
		assert.True(t, result4.IsErr())
		assert.Contains(t, result4.UnwrapErr().Error(), "conversion error")
	})

	t.Run("AndThen function", func(t *testing.T) {
		// Test AndThen with Ok result
		okResult := OK(5)
		doubledResult := AndThen(okResult, func(value int) Result[int, error] {
			return OK(value * 2)
		})
		assert.True(t, doubledResult.IsOk())
		assert.Equal(t, 10, doubledResult.Unwrap())

		// Test AndThen with Err result
		errResult := Err[int, error](errors.New("original error"))
		processedResult := AndThen(errResult, func(value int) Result[int, error] {
			return OK(value * 2)
		})
		assert.True(t, processedResult.IsErr())
		assert.Contains(t, processedResult.UnwrapErr().Error(), "original error")
	})
}
