package irr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	theIrr, innerIrr   *BasicIrr
	sourceErr, rootErr error
)

type testLogger struct{ ret string }

func (l *testLogger) Warn(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}

func (l *testLogger) Error(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}

func (l *testLogger) Fatal(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}

func init() {
	theIrr = newBasicIrr("test err %d", 1)
	innerIrr = newBasicIrr("inner error")
	theIrr.inner = innerIrr
	rootErr = errors.New("root error")
	sourceErr = fmt.Errorf("source error => %w", rootErr)
	innerIrr.inner = sourceErr
}

func TestIrrToString(t *testing.T) {
	str := theIrr.ToString(false, "=the=split=")
	assert.Equal(t, "test err 1=the=split=inner error=the=split=source error => root error", str, "")
}

func TestIrrFmtV(t *testing.T) {
	str := fmt.Sprintf("%v", theIrr)
	assert.Equal(t, "test err 1, inner error, source error => root error", str, "")
	str = fmt.Sprintf("%+v", theIrr)
	assert.Equal(t, "test err 1, inner error, source error => root error", str, "")
	str = fmt.Sprintf("%+q", theIrr)
	assert.Equal(t, "\"test err 1, inner error, source error => root error\"", str, "")
}

func TestIrrError(t *testing.T) {
	str := theIrr.Error()
	assert.Equal(t, "test err 1, inner error, source error => root error", str, "")
}

func TestIrrRoot(t *testing.T) {
	assert.Equal(t, rootErr, theIrr.Root(), "root error are not equal")
}

func TestIrrSource(t *testing.T) {
	assert.Equal(t, sourceErr, theIrr.Source(), "source error are not equal")
}

func TestUnwrap(t *testing.T) {
	assert.Equal(t, innerIrr, theIrr.Unwrap(), "unwrap to innerIrr are not correct")
	assert.Equal(t, sourceErr, innerIrr.Unwrap(), "unwrap to sourceErr are not correct")
}

func TestIrrTraverseToSourceStack(t *testing.T) {
	stack := []error{
		theIrr, innerIrr, sourceErr,
	}
	_ = theIrr.TraverseToSource(func(err error, isSource bool) error {
		pop := stack[0]
		stack = stack[1:]
		assert.Equal(t, pop, err, "wrong error stack")
		if isSource {
			assert.Equal(t, 0, len(stack), "stack should finished")
		}
		return nil
	})
	assert.Equal(t, 0, len(stack), "stack should finished")
}

func TestIrrTraverseToSourceThrownErr(t *testing.T) {
	previousErr := errors.New("the previous error")
	returnedErr := errors.New("the returned error")
	err := theIrr.TraverseToSource(func(err error, isSource bool) error {
		if isSource {
			return returnedErr
		}
		return previousErr
	})
	assert.Equal(t, returnedErr, err, "return error stack")

	err = theIrr.TraverseToSource(func(err error, isSource bool) error {
		if isSource {
			return returnedErr
		}
		panic(previousErr)
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToSource(func(err error, isSource bool) error {
		panic("some error string")
	})
	assert.Exactly(t, true, errors.Is(err, ErrUntypedExecutionFailure), "should returns ErrUntypedExecutionFailure")
}

func TestIrrTraverseToRootStack(t *testing.T) {
	stack := []error{
		theIrr, innerIrr, sourceErr, rootErr,
	}
	_ = theIrr.TraverseToRoot(func(err error) error {
		pop := stack[0]
		stack = stack[1:]
		assert.Equal(t, pop, err, "wrong error stack")
		return nil
	})
	assert.Equal(t, 0, len(stack), "stack should finished")
}

func TestIrrTraverseToRootThrownErr(t *testing.T) {
	previousErr := errors.New("the previous error")
	err := theIrr.TraverseToRoot(func(error) error {
		return previousErr
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToRoot(func(err error) error {
		panic(previousErr)
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToRoot(func(err error) error {
		panic("some error string")
	})
	assert.Exactly(t, true, errors.Is(err, ErrUntypedExecutionFailure), "should returns ErrUntypedExecutionFailure")
}

func TestLog(t *testing.T) {
	l := &testLogger{}
	ir := theIrr.LogWarn(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogWarn failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogWarn")

	ir = theIrr.LogError(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogError failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogError")

	ir = theIrr.LogFatal(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogFatal failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogFatal")
}

// === 新错误码API测试 ===

// TestNewErrorCodeAPI_NearestCode 测试NearestCode方法
func TestNewErrorCodeAPI_NearestCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected int64
	}{
		{
			name: "单个错误有错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: 404,
		},
		{
			name: "单个错误无错误码",
			setup: func() IRR {
				return Error("test error")
			},
			expected: 0,
		},
		{
			name: "错误链-最外层有错误码",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error").SetCode(400)
			},
			expected: 400, // 最近的（最外层）错误码
		},
		{
			name: "错误链-只有内层有错误码",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层没有设置错误码
			},
			expected: 500, // 内层的错误码
		},
		{
			name: "错误链-多层嵌套",
			setup: func() IRR {
				root := Error("root error").SetCode(404)
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error").SetCode(400)
			},
			expected: 400, // 最外层的错误码
		},
		{
			name: "错误链-中间层有错误码",
			setup: func() IRR {
				root := Error("root error") // 无错误码
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error") // 无错误码
			},
			expected: 500, // 中间层的错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.NearestCode())
		})
	}
}

// TestNewErrorCodeAPI_CurrentCode 测试CurrentCode方法
func TestNewErrorCodeAPI_CurrentCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected int64
	}{
		{
			name: "设置了错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: 404,
		},
		{
			name: "未设置错误码",
			setup: func() IRR {
				return Error("test error")
			},
			expected: 0,
		},
		{
			name: "设置错误码为0",
			setup: func() IRR {
				return Error("test error").SetCode(0)
			},
			expected: 0,
		},
		{
			name: "错误链-只检查当前层",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层未设置错误码
			},
			expected: 0, // 当前层（外层）没有错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.CurrentCode())
		})
	}
}

// TestNewErrorCodeAPI_RootCode 测试RootCode方法
func TestNewErrorCodeAPI_RootCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected int64
	}{
		{
			name: "单个错误",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: 404,
		},
		{
			name: "错误链-根错误有错误码",
			setup: func() IRR {
				root := Error("root error").SetCode(404)
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error").SetCode(400)
			},
			expected: 404, // 根错误的错误码
		},
		{
			name: "错误链-根错误无错误码",
			setup: func() IRR {
				root := Error("root error") // 无错误码
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error").SetCode(400)
			},
			expected: 0, // 根错误没有错误码
		},
		{
			name: "包装标准错误",
			setup: func() IRR {
				stdErr := errors.New("standard error")
				return Wrap(stdErr, "wrapped error").SetCode(500)
			},
			expected: 0, // 标准错误没有错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.RootCode())
		})
	}
}

// TestNewErrorCodeAPI_HasCurrentCode 测试HasCurrentCode方法
func TestNewErrorCodeAPI_HasCurrentCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected bool
	}{
		{
			name: "未设置错误码",
			setup: func() IRR {
				return Error("test error")
			},
			expected: false,
		},
		{
			name: "设置了非零错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: true,
		},
		{
			name: "显式设置错误码为0",
			setup: func() IRR {
				return Error("test error").SetCode(0)
			},
			expected: true, // 显式设置了，即使是0
		},
		{
			name: "错误链-外层未设置",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层未设置
			},
			expected: false, // 当前层（外层）未设置
		},
		{
			name: "错误链-外层设置了",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error").SetCode(400)
			},
			expected: true, // 当前层（外层）设置了
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.HasCurrentCode())
		})
	}
}

// TestNewErrorCodeAPI_HasAnyCode 测试HasAnyCode方法
func TestNewErrorCodeAPI_HasAnyCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected bool
	}{
		{
			name: "完全没有错误码",
			setup: func() IRR {
				root := Error("root error")
				return Wrap(root, "outer error")
			},
			expected: false,
		},
		{
			name: "有错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: true,
		},
		{
			name: "错误链中有错误码",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层无错误码
			},
			expected: true, // 内层有错误码
		},
		{
			name: "所有错误码都是0",
			setup: func() IRR {
				inner := Error("inner error").SetCode(0)
				return Wrap(inner, "outer error").SetCode(0)
			},
			expected: false, // 所有错误码都是0，视为无有效错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.HasAnyCode())
		})
	}
}

// TestNewErrorCodeAPI_BackwardCompatibility 测试向后兼容性
func TestNewErrorCodeAPI_BackwardCompatibility(t *testing.T) {
	// 创建一个错误链
	inner := Error("inner error").SetCode(404)
	wrapped := Wrap(inner, "wrapped error").SetCode(500)

	// 测试GetCode和NearestCode返回相同结果
	assert.Equal(t, wrapped.NearestCode(), wrapped.GetCode())
	
	// 测试ClosestCode和NearestCode返回相同结果
	assert.Equal(t, wrapped.NearestCode(), wrapped.ClosestCode())

	// 测试具体值
	assert.Equal(t, int64(500), wrapped.GetCode())
	assert.Equal(t, int64(500), wrapped.ClosestCode())
	assert.Equal(t, int64(500), wrapped.NearestCode())
}

// TestNewErrorCodeAPI_SetCodeBehavior 测试SetCode的新行为
func TestNewErrorCodeAPI_SetCodeBehavior(t *testing.T) {
	err := Error("test error")
	
	// 初始状态
	assert.False(t, err.HasCurrentCode())
	assert.Equal(t, int64(0), err.CurrentCode())
	
	// 设置错误码
	err.SetCode(404)
	assert.True(t, err.HasCurrentCode())
	assert.Equal(t, int64(404), err.CurrentCode())
	
	// 设置错误码为0
	err.SetCode(0)
	assert.True(t, err.HasCurrentCode()) // 仍然是true，因为显式设置了
	assert.Equal(t, int64(0), err.CurrentCode())
}

// TestNewErrorCodeAPI_ComplexChain 测试复杂错误链
func TestNewErrorCodeAPI_ComplexChain(t *testing.T) {
	// 创建复杂的错误链
	stdErr := errors.New("standard error")
	level1 := Wrap(stdErr, "level 1").SetCode(100)
	level2 := Wrap(level1, "level 2") // 无错误码
	level3 := Wrap(level2, "level 3").SetCode(300)
	level4 := Wrap(level3, "level 4") // 无错误码
	level5 := Wrap(level4, "level 5").SetCode(500)

	// 测试各种方法
	assert.Equal(t, int64(500), level5.NearestCode()) // 最近的有效错误码
	assert.Equal(t, int64(500), level5.CurrentCode()) // 当前层错误码
	assert.Equal(t, int64(0), level5.RootCode())      // 根错误（标准错误）无错误码
	assert.True(t, level5.HasCurrentCode())           // 当前层设置了错误码
	assert.True(t, level5.HasAnyCode())               // 错误链中有错误码

	// 测试中间层
	assert.Equal(t, int64(300), level4.NearestCode()) // 最近的有效错误码是level3的
	assert.Equal(t, int64(0), level4.CurrentCode())   // 当前层无错误码
	assert.False(t, level4.HasCurrentCode())          // 当前层未设置错误码
	assert.True(t, level4.HasAnyCode())               // 错误链中有错误码
}

// TestNewErrorCodeAPI_EdgeCases 测试边界情况
func TestNewErrorCodeAPI_EdgeCases(t *testing.T) {
	t.Run("空错误链", func(t *testing.T) {
		err := Error("")
		assert.Equal(t, int64(0), err.NearestCode())
		assert.Equal(t, int64(0), err.CurrentCode())
		assert.Equal(t, int64(0), err.RootCode())
		assert.False(t, err.HasCurrentCode())
		assert.False(t, err.HasAnyCode())
	})

	t.Run("多次设置错误码", func(t *testing.T) {
		err := Error("test")
		err.SetCode(100)
		err.SetCode(200)
		err.SetCode(300)
		
		assert.Equal(t, int64(300), err.CurrentCode())
		assert.True(t, err.HasCurrentCode())
	})

	t.Run("负数错误码", func(t *testing.T) {
		err := Error("test").SetCode(-1)
		assert.Equal(t, int64(-1), err.CurrentCode())
		assert.Equal(t, int64(-1), err.NearestCode())
		assert.True(t, err.HasCurrentCode())
		assert.True(t, err.HasAnyCode())
	})
}
