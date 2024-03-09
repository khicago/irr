package result

func OK[T any](result T) Result[T] {
	return Result[T]{
		result: result,
	}
}

func Err[T any](err error) Result[T] {
	return Result[T]{
		err: err,
	}
}

// AndThen - 处理链
// 由于Go 不支持方法形参，因此用包函数的方式提供
// 这个方法将接收一个闭包，如果 Result 是成功的，它会调用闭包，其余上下文直接在闭包内携带，比如 ctx
func AndThen[T any, U any](r Result[T], op func(T) Result[U]) Result[U] {
	if !r.Ok() {
		// 如果当前 Result 有错误，直接返回一个带该错误的新 Result
		return Result[U]{err: r.err}
	}
	// 否则，调用 op 传入成功的值，并返回 op 调用的结果
	// r 不适用闭包，而是直接传入，主要是为了清晰的作用域，避免闭包内的 unwrap
	return op(r.result)
}
