package errors

// traversal.go - 错误链遍历和导航相关功能
// 包含错误链的各种遍历操作和根错误查找

// ========== 错误链遍历 ==========

// Root 获取根错误
func (e *ErrorImpl) Root() error {
	current := error(e)
	for {
		if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			if next := unwrapper.Unwrap(); next != nil {
				current = next
				continue
			}
		}
		break
	}
	return current
}

// TraverseToRoot 遍历到根错误
func (e *ErrorImpl) TraverseToRoot(fn func(err error) error) error {
	current := error(e)
	for current != nil {
		if result := fn(current); result != nil {
			return result
		}

		if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			current = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return nil
}

// TraverseToSource 遍历到源错误
func (e *ErrorImpl) TraverseToSource(fn func(err error, isSource bool) error) error {
	current := error(e)

	for current != nil {
		var next error
		if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			next = unwrapper.Unwrap()
		}

		isSource := (next == nil) // 是源错误（最底层）
		if result := fn(current, isSource); result != nil {
			return result
		}

		current = next
	}
	return nil
}
