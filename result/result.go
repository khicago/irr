package result

import (
	"github.com/khicago/irr"
)

// Result 标准的Result<T,E>模式实现
// 基于第一性原理，提供类型安全的错误处理
type Result[T, E any] struct {
	value T
	err   E
	isOk  bool
}

// Go 1.23+ 泛型别名 - 零开销抽象
type ResultIRR[T any] = Result[T, irr.IRR]
type ResultError[T any] = Result[T, error]

// ========== 构造函数 ==========

// Ok 创建成功的Result
func Ok[T, E any](value T) Result[T, E] {
	return Result[T, E]{
		value: value,
		isOk:  true,
	}
}

// Err 创建失败的Result
func Err[T, E any](err E) Result[T, E] {
	return Result[T, E]{
		err:  err,
		isOk: false,
	}
}

// ========== 语法糖构造函数 ==========

// OkIRR 创建成功的ResultIRR[T]
func OkIRR[T any](value T) ResultIRR[T] {
	return Ok[T, irr.IRR](value)
}

// ErrIRR 创建失败的ResultIRR[T]
func ErrIRR[T any](err irr.IRR) ResultIRR[T] {
	return Err[T, irr.IRR](err)
}

// OkError 创建成功的ResultError[T]
func OkError[T any](value T) ResultError[T] {
	return Ok[T, error](value)
}

// ErrError 创建失败的ResultError[T]
func ErrError[T any](err error) ResultError[T] {
	return Err[T, error](err)
}

// ========== 便利别名 ==========

// OK 创建成功的Result（便利别名，兼容现有代码）
func OK[T any](value T) ResultError[T] {
	return OkError(value)
}

// ========== 状态检查 ==========

// IsOk 检查是否成功
func (r Result[T, E]) IsOk() bool {
	return r.isOk
}

// IsErr 检查是否失败
func (r Result[T, E]) IsErr() bool {
	return !r.isOk
}

// Ok 检查是否成功（简短API）
func (r Result[T, E]) Ok() bool {
	return r.isOk
}

// Err 获取错误值（简短API）
func (r Result[T, E]) Err() E {
	return r.err
}

// ========== 值提取 ==========

// Value 安全提取值
func (r Result[T, E]) Value() (T, bool) {
	if r.isOk {
		return r.value, true
	}
	var zero T
	return zero, false
}

// Error 安全提取错误
func (r Result[T, E]) Error() (E, bool) {
	if !r.isOk {
		return r.err, true
	}
	var zero E
	return zero, false
}

// Unwrap 强制提取值（可能panic）
func (r Result[T, E]) Unwrap() T {
	if !r.isOk {
		panic("called Unwrap on Err Result")
	}
	return r.value
}

// UnwrapOr 提取值或返回默认值
func (r Result[T, E]) UnwrapOr(defaultVal T) T {
	if r.isOk {
		return r.value
	}
	return defaultVal
}

// UnwrapErr 强制提取错误（可能panic）
func (r Result[T, E]) UnwrapErr() E {
	if r.isOk {
		panic("called UnwrapErr on Ok Result")
	}
	return r.err
}

// Expect 强制提取值，失败时panic带自定义消息
func (r Result[T, E]) Expect(msg string) T {
	if !r.isOk {
		panic(msg)
	}
	return r.value
}

// ========== 模式匹配 ==========

// Match 模式匹配处理，返回(value, error)元组
func (r Result[T, E]) Match() (T, E) {
	if r.isOk {
		var zero E
		return r.value, zero
	} else {
		var zero T
		return zero, r.err
	}
}

// MatchFunc 基于函数的模式匹配处理
func (r Result[T, E]) MatchFunc(
	onOk func(T),
	onErr func(E),
) {
	if r.isOk {
		onOk(r.value)
	} else {
		onErr(r.err)
	}
}

// ========== 标准库集成 ==========

// ToTuple 转换为Go惯用的(value, error)元组
func ToTuple[T any](r ResultError[T]) (T, error) {
	if r.isOk {
		return r.value, nil
	}
	var zero T
	return zero, r.err
}

// FromTuple 从(value, error)元组创建Result
func FromTuple[T any](value T, err error) ResultError[T] {
	if err != nil {
		return ErrError[T](err)
	}
	return OkError(value)
}
