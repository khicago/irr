package irr

import (
	"errors"
)

type (
	ILogCaller interface {
		LogWarn(logger IWarnLogger) IRR
		LogError(logger IErrorLogger) IRR
		LogFatal(logger IFatalLogger) IRR
	}
)

type (
	ICoder[TCode any] interface {
		ICodeGetter[TCode]
		SetCode(val TCode) IRR
	}

	ICodeGetter[TCode any] interface {
		GetCode() (val TCode)
		GetCodeStr() string
	}

	ITagger interface {
		SetTag(key, value string)
		GetTag(key string) (values []string)
	}
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
		ICodeGetter[TCode]
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
