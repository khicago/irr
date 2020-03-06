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

		GetTraceInfo() *traceInfo

		TraverseToSource(fn func(err error) error) (err error)
		TraverseToRoot(fn func(err error) error) (err error)

		LogWarn(logger interface{ Warn(args ...interface{}) }) IRR
		LogError(logger interface{ Error(args ...interface{}) }) IRR
		LogFatal(logger interface{ Fatal(args ...interface{}) }) IRR
	}
)

var _ IRR = &irr{}

var (
	ErrUntypedExecutionFailure = errors.New("untyped execution failure")
)
