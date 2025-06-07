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

	// 新的清晰错误码API接口
	ICodeManager[TCode any] interface {
		// 错误码获取
		NearestCode() TCode // 最近的有效错误码
		CurrentCode() TCode // 当前对象的错误码
		RootCode() TCode    // 根错误的错误码

		// 错误码状态检查
		HasCurrentCode() bool // 当前对象是否设置了错误码
		HasAnyCode() bool     // 错误链中是否有任何错误码

		// 向后兼容（标记为废弃）
		GetCode() TCode     // @deprecated: 使用 NearestCode()
		ClosestCode() TCode // @deprecated: 使用 NearestCode()
	}

	IRR interface {
		ITraverseIrr

		error
		ITraverseError
		IUnwrap

		ICodeManager[int64]    // 使用新的清晰错误码API
		SetCode(val int64) IRR // 保留SetCode方法
		GetCodeStr() string    // 保留GetCodeStr方法

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
