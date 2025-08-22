package result

import (
	"github.com/khicago/irr"
)

// ========== 单类型参数便利函数 ==========

// 为了支持Result[T]语法（等同于Result[T, IRR]），提供便利函数
// 基于第一性原理，让最常用的API最简洁

// OkResult 创建单类型参数的成功Result[T] (等同于Result[T, IRR])
func OkResult[T any](value T) ResultIRR[T] {
	return OkIRR(value)
}

// ErrResult 创建单类型参数的失败Result[T] (等同于Result[T, IRR])
func ErrResult[T any](err irr.IRR) ResultIRR[T] {
	return ErrIRR[T](err)
}

// ========== 高级函数式操作 ==========

// AndThen 链式操作：如果当前Result是Ok，则执行操作并返回新Result
// 这是Monad模式的核心操作，支持错误短路
func AndThen[T, U, E any](r Result[T, E], op func(T) Result[U, E]) Result[U, E] {
	if !r.IsOk() {
		// 错误短路：直接传播错误，不执行操作
		return Err[U, E](r.err)
	}
	// 执行操作，返回新的Result
	return op(r.value)
}

// OrElse 错误恢复：如果当前Result是Err，则执行恢复操作
func OrElse[T, E any](r Result[T, E], recovery func(E) Result[T, E]) Result[T, E] {
	if r.IsOk() {
		return r // 成功时直接返回
	}
	// 尝试从错误中恢复
	return recovery(r.err)
}

// Map 函数式映射：转换Result[T,E]为Result[U,E]
// 只对成功值进行转换，错误直接传播
func Map[T, U, E any](r Result[T, E], fn func(T) U) Result[U, E] {
	if r.IsOk() {
		return Ok[U, E](fn(r.value))
	}
	return Err[U, E](r.err)
}

// MapErr 错误映射：转换Result[T,E]为Result[T,F]
// 只对错误值进行转换，成功值直接传播
func MapErr[T, E, F any](r Result[T, E], fn func(E) F) Result[T, F] {
	if r.IsOk() {
		return Ok[T, F](r.value)
	}
	return Err[T, F](fn(r.err))
}

// Collect 将多个Result收集为一个包含切片的Result
// 如果任何一个Result是错误，则返回第一个错误
func Collect[T, E any](results ...Result[T, E]) Result[[]T, E] {
	values := make([]T, 0, len(results))
	for _, r := range results {
		if r.IsErr() {
			return Err[[]T, E](r.err)
		}
		values = append(values, r.value)
	}
	return Ok[[]T, E](values)
}

// Flatten 扁平化嵌套的Result
func Flatten[T, E any](r Result[Result[T, E], E]) Result[T, E] {
	if r.IsErr() {
		return Err[T, E](r.err)
	}
	return r.value
}

// Filter 过滤Result中的值
func Filter[T, E any](r Result[T, E], predicate func(T) bool, onFilterFail func() E) Result[T, E] {
	if r.IsErr() {
		return r
	}
	if predicate(r.value) {
		return r
	}
	return Err[T, E](onFilterFail())
}

// ========== 实用工具函数 ==========

// IsOkAnd 检查是否成功且值满足条件
func IsOkAnd[T, E any](r Result[T, E], predicate func(T) bool) bool {
	if !r.IsOk() {
		return false
	}
	return predicate(r.value)
}

// IsErrAnd 检查是否失败且错误满足条件
func IsErrAnd[T, E any](r Result[T, E], predicate func(E) bool) bool {
	if !r.IsErr() {
		return false
	}
	return predicate(r.err)
}

// ========== ResultIRR专用工具 ==========

// ExpectIRR 期望成功值，如果是IRR错误则包含详细信息
func ExpectIRR[T any](r ResultIRR[T], msg string) T {
	if !r.IsOk() {
		if err, ok := r.Error(); ok {
			panic(msg + " - " + err.Details())
		}
		panic(msg + " - Result was Err")
	}
	return r.value
}

// ToIRRTuple 将ResultIRR转换为(value, IRR)元组
func ToIRRTuple[T any](r ResultIRR[T]) (T, irr.IRR) {
	if r.IsOk() {
		return r.value, nil
	}
	var zero T
	return zero, r.err
}

// FromIRRTuple 从(value, IRR)元组创建ResultIRR
func FromIRRTuple[T any](value T, err irr.IRR) ResultIRR[T] {
	if err != nil {
		return ErrIRR[T](err)
	}
	return OkIRR(value)
}
