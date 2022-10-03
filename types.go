package irr

import (
	"errors"
)

type (
	IUnwrap interface {
		Unwrap() error
	}

	ITraverseError interface {
		Root() error
		TraverseToRoot(fn func(err error) error) (err error)
	}

	ITraverseIrr interface {
		Source() error
		TraverseToSource(fn func(err error, isSource bool) error) (err error)
	}

	ITraverseCoder[TCode any] interface {
		ClosestCode() TCode
		TraverseCode(fn func(err error, code TCode) error) (err error)
	}

	ILogCaller interface {
		LogWarn(logger IWarnLogger) IRR
		LogError(logger IErrorLogger) IRR
		LogFatal(logger IFatalLogger) IRR
	}

	ICoder[TCode any] interface {
		SetCode(val TCode) IRR
		GetCode() (val TCode)
	}

	ITagger interface {
		SetTag(key, value string)
		GetTag(key string) (values []string)
	}

	IRR interface {
		ITraverseIrr

		error
		ITraverseError
		IUnwrap

		ICoder[int64]
		ITraverseCoder[int64]

		ITagger
		ILogCaller

		ToString(printTrace bool, split string) string
		GetTraceInfo() *traceInfo
	}

	Spawner interface {
		Error(formatOrMsg string, args ...interface{}) IRR
		Wrap(innerErr error, formatOrMsg string, args ...interface{}) IRR
		TraceSkip(skip int, formatOrMsg string, args ...interface{}) IRR
		Trace(formatOrMsg string, args ...interface{}) IRR
		TrackSkip(skip int, innerErr error, formatOrMsg string, args ...interface{}) IRR
		Track(innerErr error, formatOrMsg string, args ...interface{}) IRR
	}
)

var ErrUntypedExecutionFailure = errors.New("!!!panic")
