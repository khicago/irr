package irr

import (
	"context"
	"fmt"
)

// IRR - 基于第一性原理的最终错误接口
// 保留有用功能，删除冗余兼容代码
type IRR interface {
	// 基础错误接口
	error
	fmt.Formatter

	// 错误链操作
	Unwrap() error
	Root() error

	// 错误码管理 - 统一API
	Code() int64
	HasCode() bool
	SetCode(code int64) IRR

	// 格式化输出
	String() string
	Details() string
	ToString(printTrace bool, split string) string // 兼容现有代码的格式化方法

	// 堆栈跟踪
	StackTrace() []Frame
	HasStackTrace() bool

	// 标签支持
	Tags() map[string][]string
	SetTag(key, value string)
	Tag(key, value string) IRR
	GetTag(key string) []string // 获取特定标签的值

	// 上下文感知
	Context() context.Context
	WithContext(ctx context.Context) IRR

	// 错误遍历 - 保留有用的遍历功能
	TraverseToRoot(fn func(err error) error) error
	TraverseToSource(fn func(err error, isSource bool) error) error

	// 注意: 删除了冗余的GetCode()、GetTraceInfo()、ClosestCode()方法
	// 统一使用Code()和StackTrace()
}

// Spawner 错误生成器接口
type Spawner interface {
	Error(message string, args ...any) IRR
	Wrap(err error, message string, args ...any) IRR
	Track(err error, message string, args ...any) IRR
	Trace(message string, args ...any) IRR
}

// ICoder 编码器接口，支持泛型错误码类型
type ICoder[T any] interface {
	Code() T
	HasCode() bool
	// 注意: 删除了冗余的GetCode()方法，统一使用Code()
}
