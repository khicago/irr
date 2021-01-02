package irr

import (
	"errors"
)

type (
	irr struct {
		inner error

		Msg   string     `json:"msg"`
		Trace *traceInfo `json:"trace"`
	}

	IRR interface {
		error

		Root() error
		Source() error
		Unwrap() error

		TraverseToSource(fn func(err error, isSource bool) error) (err error)
		TraverseToRoot(fn func(err error) error) (err error)

		LogWarn(logger interface{ Warn(args ...interface{}) }) IRR
		LogError(logger interface{ Error(args ...interface{}) }) IRR
		LogFatal(logger interface{ Fatal(args ...interface{}) }) IRR

		ToString(printTrace bool, split string) string
		GetTraceInfo() *traceInfo
	}
)

var _ IRR = &irr{}

var (
	ErrUntypedExecutionFailure = errors.New("untyped execution failure")
)
