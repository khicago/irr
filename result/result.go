package result

import (
	"github.com/khicago/irr"
)

var ErrUnwrapErrOnOK = irr.Error("called UnwrapErr on ok")

type Result[T any] struct {
	result T
	err    error
}

func (r Result[T]) Ok() bool {
	return r.err == nil
}

func (r Result[T]) Err() error {
	return r.err
}

// Match 获取存储在 Result 中的值，如果存在错误，则返回默认值和错误
//
// switch res, err := r.Match() {
// case err != nil: error handling branch
// default: ...
// }
func (r Result[T]) Match() (T, error) {
	if r.err != nil {
		var zero T
		return zero, r.err
	}
	return r.result, nil
}

// Unwrap 强制解包 Result.result，如果 Result 包含错误，则抛出 panic
// Result 不会被消耗，todo 这个可以考虑考虑
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(r.Err)
	}
	return r.result
}

// UnwrapOr 强制解包 Result，如果 Result 不包含错误，则抛出默认值
// Result 不会被消耗，todo 这个可以考虑考虑
func (r Result[T]) UnwrapOr(defaultVal T) T {
	if r.err != nil {
		return defaultVal
	}
	return r.result
}

// UnwrapErr 强制解包 Result.Err，如果 Result 不包含错误，则抛出 panic
// Result 不会被消耗，todo 这个可以考虑考虑
func (r Result[T]) UnwrapErr() error {
	if r.err == nil {
		panic(irr.Wrap(ErrUnwrapErrOnOK, "value= %v", r.result))
	}
	return r.err
}

// Expect 返回 Result 中的值或者在发生错误时显示指定的消息
// Result 不会被消耗，todo 这个可以考虑考虑
func (r Result[T]) Expect(formatOrMsg string, params ...any) T {
	if r.err != nil {
		e := irr.Wrap(r.Err(), formatOrMsg, params...)
		panic(e)
	}
	return r.result
}
