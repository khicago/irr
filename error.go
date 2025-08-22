package irr

import (
	"context"
	"fmt"

	"github.com/khicago/irr/internal/errors"
)

// Frame 堆栈跟踪帧信息 - 重新导出内部类型
type Frame = errors.Frame

// errorImpl 实现IRR接口的具体类型
// 基于internal/errors.ErrorImpl，提供完整的IRR接口实现
type errorImpl struct {
	*errors.ErrorImpl
}

// 确保 errorImpl 实现 IRR 接口
var _ IRR = (*errorImpl)(nil)

// ========== IRR 接口适配方法 ==========

// SetCode 设置错误码 - 返回IRR接口
func (e *errorImpl) SetCode(code int64) IRR {
	e.ErrorImpl.SetCode(code)
	return e
}

// Tag 设置标签 - 返回IRR接口
func (e *errorImpl) Tag(key, value string) IRR {
	e.ErrorImpl.Tag(key, value)
	return e
}

// WithContext 设置上下文 - 返回IRR接口
func (e *errorImpl) WithContext(ctx context.Context) IRR {
	e.ErrorImpl.WithContext(ctx)
	return e
}

// Root 获取根错误 - 确保返回类型一致
func (e *errorImpl) Root() error {
	root := e.ErrorImpl.Root()
	// 如果根错误就是自己，返回自己
	if root == e.ErrorImpl {
		return e
	}
	return root
}

// ========== 构造函数 ==========

// newErrorImpl 创建新的errorImpl实例 - 使用内部实现
func newErrorImpl(message string, code int64, cause error, tags map[string][]string, trace []Frame, ctx context.Context) *errorImpl {
	return &errorImpl{
		ErrorImpl: errors.NewErrorImpl(message, code, cause, tags, trace, ctx),
	}
}

// ========== 格式化接口桥接 ==========

// String 基础字符串表示
func (e *errorImpl) String() string {
	return e.ErrorImpl.String()
}

// Error 标准错误接口
func (e *errorImpl) Error() string {
	return e.ErrorImpl.Error()
}

// Format 格式化接口
func (e *errorImpl) Format(s fmt.State, verb rune) {
	e.ErrorImpl.Format(s, verb)
}
