package irr

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	theIrr, innerIrr   *irr
	sourceErr, rootErr error
)

type testLogger struct{ ret string }

func (l *testLogger) Warn(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}
func (l *testLogger) Error(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}
func (l *testLogger) Fatal(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}

func init() {
	theIrr = newIrr("test err %d", 1)
	innerIrr = newIrr("inner error")
	theIrr.inner = innerIrr
	rootErr = errors.New("root error")
	sourceErr = fmt.Errorf("source error => %w", rootErr)
	innerIrr.inner = sourceErr
}

func TestIrrToString(t *testing.T) {
	str := theIrr.ToString(false, "=the=split=")
	assert.Equal(t, "test err 1=the=split=inner error=the=split=source error => root error", str, "")
}

func TestIrrError(t *testing.T) {
	str := theIrr.Error()
	assert.Equal(t, "test err 1; inner error; source error => root error", str, "")
}

func TestIrrRoot(t *testing.T) {
	assert.Equal(t, rootErr, theIrr.Root(), "root error are not equal")
}

func TestIrrSource(t *testing.T) {
	assert.Equal(t, sourceErr, theIrr.Source(), "source error are not equal")
}

func TestUnwrap(t *testing.T) {
	assert.Equal(t, innerIrr, theIrr.Unwrap(), "unwrap to innerIrr are not correct")
	assert.Equal(t, sourceErr, innerIrr.Unwrap(), "unwrap to sourceErr are not correct")
}

func TestIrrTraverseToSourceStack(t *testing.T) {
	stack := []error{
		theIrr, innerIrr, sourceErr,
	}
	_ = theIrr.TraverseToSource(func(err error, isSource bool) error {
		pop := stack[0]
		stack = stack[1:]
		assert.Equal(t, pop, err, "wrong error stack")
		if isSource {
			assert.Equal(t, 0, len(stack), "stack should finished")
		}
		return nil
	})
	assert.Equal(t, 0, len(stack), "stack should finished")
}

func TestIrrTraverseToSourceThrownErr(t *testing.T) {
	previousErr := errors.New("the previous error")
	returnedErr := errors.New("the returned error")
	err := theIrr.TraverseToSource(func(err error, isSource bool) error {
		if isSource {
			return returnedErr
		}
		return previousErr
	})
	assert.Equal(t, returnedErr, err, "return error stack")

	err = theIrr.TraverseToSource(func(err error, isSource bool) error {
		if isSource {
			return returnedErr
		}
		panic(previousErr)
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToSource(func(err error, isSource bool) error {
		panic("some error string")
	})
	assert.Exactly(t, true, errors.Is(err, ErrUntypedExecutionFailure), "should returns ErrUntypedExecutionFailure")
}

func TestIrrTraverseToRootStack(t *testing.T) {
	stack := []error{
		theIrr, innerIrr, sourceErr, rootErr,
	}
	_ = theIrr.TraverseToRoot(func(err error) error {
		pop := stack[0]
		stack = stack[1:]
		assert.Equal(t, pop, err, "wrong error stack")
		return nil
	})
	assert.Equal(t, 0, len(stack), "stack should finished")
}

func TestIrrTraverseToRootThrownErr(t *testing.T) {
	previousErr := errors.New("the previous error")
	returnedErr := errors.New("the returned error")
	err := theIrr.TraverseToRoot(func(err error) error {
		pe := previousErr
		previousErr = returnedErr
		return pe
	})
	assert.Equal(t, returnedErr, err, "return error stack")

	err = theIrr.TraverseToRoot(func(err error) error {
		pe := previousErr
		previousErr = returnedErr
		panic(pe)
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToRoot(func(err error) error {
		panic("some error string")
	})
	assert.Exactly(t, true, errors.Is(err, ErrUntypedExecutionFailure), "should returns ErrUntypedExecutionFailure")
}

func TestLog(t *testing.T) {
	l := &testLogger{}
	ir := theIrr.LogWarn(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogWarn failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogWarn")

	ir = theIrr.LogError(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogError failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogError")

	ir = theIrr.LogFatal(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogFatal failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogFatal")
}
