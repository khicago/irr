package irr

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

type (
	BasicIrr struct {
		inner error

		Code    int64      `json:"code"`
		codeSet bool       `json:"code_set"` // 跟踪是否显式设置过错误码
		Msg     string     `json:"msg"`
		Trace   *traceInfo `json:"trace"`

		// 使用 map 替代 slice，提升查找性能
		// 使用原子操作的指针，减少锁竞争
		tags atomic.Pointer[map[string][]string] `json:"-"`
		mu   sync.RWMutex                        // 保留锁用于tag操作的原子性
	}
)

func newBasicIrr(formatOrMsg string, args ...any) *BasicIrr {
	err := &BasicIrr{}
	if len(args) > 0 {
		err.Msg = fmt.Sprintf(formatOrMsg, args...)
	} else {
		err.Msg = formatOrMsg
	}
	return err
}

var _ IRR = newBasicIrr("")

// Error
// the implementation of error
func (ir *BasicIrr) Error() string {
	return ir.ToString(false, ", ")
}

// Unwrap
// the implementation of IUnwrap
func (ir *BasicIrr) Unwrap() error {
	return ir.inner
}

// Root
// the implementation of ITraverseError
func (ir *BasicIrr) Root() error {
	var err error = ir
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			return err
		}
		err = inner
	}
}

// TraverseToRoot
// the implementation of ITraverseError
func (ir *BasicIrr) TraverseToRoot(fn func(err error) error) (err error) {
	recordTraverseOp()
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = Wrap(ErrUntypedExecutionFailure, "panic = %v", r)
			}
		}
	}()
	for inner := error(ir); inner != nil; inner = errors.Unwrap(inner) {
		if err = fn(inner); err != nil {
			return err
		}
	}
	return
}

// Source
// the implementation of ITraverseIrr
func (ir *BasicIrr) Source() (err error) {
	_ = ir.TraverseToSource(func(e error, isSource bool) error {
		if isSource {
			err = e
		}
		return nil
	})
	return
}

// TraverseToSource
// the implementation of ITraverseIrr
func (ir *BasicIrr) TraverseToSource(fn func(err error, isSource bool) error) (err error) {
	recordTraverseOp()
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = Wrap(ErrUntypedExecutionFailure, "panic = %v", r)
			}
		}
	}()
	for cur := ir; cur != nil; {
		isCurSource := cur.inner == nil
		err = fn(cur, isCurSource)
		if isCurSource {
			// 只有在 source 时才返回函数的结果
			return err
		}
		if next, ok := cur.inner.(*BasicIrr); ok {
			cur = next
			continue
		}
		// 到达最后一个非 BasicIrr 的错误，这是 source
		err = fn(cur.inner, true)
		return err
	}
	return nil
}

// GetCodeStr
// Determines how the code is written to the message,
// so that this method can input an empty string to
// avoid outputting the code in the message
func (ir *BasicIrr) GetCodeStr() string {
	if ir.Code == 0 {
		return ""
	}
	return fmt.Sprintf("code(%d), ", ir.Code)
}

func (ir *BasicIrr) writeSelfTo(sb *strings.Builder, printTrace bool, printCode bool) {
	if printCode {
		if codeStr := ir.GetCodeStr(); codeStr != "" {
			sb.WriteString(codeStr)
		}
	}
	sb.WriteString(ir.Msg)

	// 获取tags进行输出
	if tagMap := ir.tags.Load(); tagMap != nil && len(*tagMap) > 0 {
		for key, values := range *tagMap {
			for _, value := range values {
				sb.WriteRune('[')
				sb.WriteString(key)
				sb.WriteRune(':')
				sb.WriteString(value)
				sb.WriteString("] ")
			}
		}
	}
	if printTrace && ir.Trace != nil {
		sb.WriteRune(' ')
		ir.Trace.writeTo(sb)
	}
}

// ToString
// consecutive equal codes will be printed only once during the traceback process
func (ir *BasicIrr) ToString(printTrace bool, split string) string {
	sb := strings.Builder{}
	lastCode := int64(0)
	_ = ir.TraverseToSource(func(err error, isSource bool) error {
		if irr, ok := err.(*BasicIrr); ok {
			// since have to continue traversing, irr only output itself
			irr.writeSelfTo(&sb, printTrace, lastCode != irr.Code)
			lastCode = irr.Code
		} else {
			sb.WriteString(err.Error())
		}
		if !isSource {
			sb.WriteString(split)
		}
		return nil
	})
	return sb.String()
}

// LogWarn
// the implementation of ILogCaller
func (ir *BasicIrr) LogWarn(logger IWarnLogger) IRR {
	logger.Warn(ir.ToString(true, "\n"))
	return ir
}

// LogError
// the implementation of ILogCaller
func (ir *BasicIrr) LogError(logger IErrorLogger) IRR {
	logger.Error(ir.ToString(true, "\n"))
	return ir
}

// LogFatal
// the implementation of ILogCaller
func (ir *BasicIrr) LogFatal(logger IFatalLogger) IRR {
	str := ir.ToString(true, "\n")
	logger.Fatal(str)
	// to make sure it has been printed to std output stream
	fmt.Println(str)
	return ir
}

// SetCode
// the implementation of SetCode method
func (ir *BasicIrr) SetCode(val int64) IRR {
	if val != 0 {
		recordErrorWithCode(val)
	}
	ir.Code = val
	ir.codeSet = true // 标记已设置错误码
	return ir
}

// === 新的清晰错误码API ===

// NearestCode 返回错误链中最近的有效错误码（非零）
// 这是推荐使用的方法，符合用户直觉
func (ir *BasicIrr) NearestCode() int64 {
	eExit := errors.New("stop")
	var val int64
	if err := ir.TraverseCode(func(_ error, code int64) error {
		if code != 0 {
			val = code
			return eExit
		}
		return nil
	}); err != nil && err != eExit {
		panic("traverse panic")
	}
	return val
}

// CurrentCode 返回当前错误对象的错误码（可能为0）
func (ir *BasicIrr) CurrentCode() int64 {
	return ir.Code
}

// RootCode 返回错误链根部的错误码
func (ir *BasicIrr) RootCode() int64 {
	var rootCode int64
	var rootErr error

	// 先获取根错误
	_ = ir.TraverseToSource(func(err error, isSource bool) error {
		if isSource {
			rootErr = err
		}
		return nil
	})

	// 检查根错误是否有错误码
	if rootErr != nil {
		if t, ok := rootErr.(*BasicIrr); ok {
			rootCode = t.Code
		} else if t, ok := rootErr.(interface{ CurrentCode() int64 }); ok {
			rootCode = t.CurrentCode()
		} else if t, ok := rootErr.(interface{ GetCode() int64 }); ok {
			// 兼容其他实现了GetCode的错误类型
			rootCode = t.GetCode()
		}
	}

	return rootCode
}

// HasCurrentCode 检查当前错误对象是否显式设置了错误码
func (ir *BasicIrr) HasCurrentCode() bool {
	return ir.codeSet
}

// HasAnyCode 检查错误链中是否有任何错误码
func (ir *BasicIrr) HasAnyCode() bool {
	return ir.NearestCode() != 0
}

// === 向后兼容的废弃方法 ===

// GetCode 返回最近的有效错误码
// Deprecated: 使用 NearestCode() 获得更清晰的语义
func (ir *BasicIrr) GetCode() int64 {
	return ir.NearestCode()
}

// ClosestCode 返回最近的有效错误码
// Deprecated: 使用 NearestCode() 获得更清晰的语义
func (ir *BasicIrr) ClosestCode() int64 {
	return ir.NearestCode()
}

// TraverseCode
// the implementation of ITraverseCoder[int64]
func (ir *BasicIrr) TraverseCode(fn func(err error, code int64) error) (err error) {
	return ir.TraverseToRoot(func(err error) error {
		var code int64
		if t, ok := err.(interface{ CurrentCode() int64 }); ok {
			code = t.CurrentCode()
		} else if t, ok := err.(interface{ GetCode() int64 }); ok {
			// 兼容其他实现了GetCode的错误类型
			code = t.GetCode()
		}
		if err = fn(err, code); err != nil {
			return err
		}
		return nil
	})
}

// SetTag
// the implementation of ITagger
func (ir *BasicIrr) SetTag(key, val string) {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	// 获取当前的tags map
	currentTags := ir.tags.Load()
	var newTags map[string][]string

	if currentTags == nil {
		newTags = make(map[string][]string)
	} else {
		// 复制现有的tags
		newTags = make(map[string][]string, len(*currentTags))
		for k, v := range *currentTags {
			newTags[k] = make([]string, len(v))
			copy(newTags[k], v)
		}
	}

	// 添加新的tag
	newTags[key] = append(newTags[key], val)
	ir.tags.Store(&newTags)
}

// GetTag
// the implementation of ITagger
func (ir *BasicIrr) GetTag(key string) (val []string) {
	tagMap := ir.tags.Load()
	if tagMap == nil {
		return nil
	}
	values := (*tagMap)[key]
	if values == nil {
		return nil
	}
	// 返回副本以避免竞态条件
	result := make([]string, len(values))
	copy(result, values)
	return result
}

func (ir *BasicIrr) GetTraceInfo() *traceInfo {
	return ir.Trace
}
